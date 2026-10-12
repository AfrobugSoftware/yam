package yui

import (
	"errors"
	"image"
	"image/draw"
	"os"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

type Font struct {
	TextureID       uint32
	VerticalSpacing float32
	buffer          []byte
	pointSize       float32
	ppi             float32
	font            *sfnt.Font
	width           int
	height          int
	Kern            float32
	ppem            fixed.Int26_6
	gmap            map[rune]Character
}

type Character struct {
	Advance float32
	Size    y3d.Vec2
	Idx     uint16
}

func NewFont(pointSize, ppi float32, filename string) *Font {
	f := &Font{
		pointSize: pointSize,
		ppi:       ppi,
	}

	gl.CreateTextures(gl.TEXTURE_2D_ARRAY, 1, &f.TextureID)
	gl.TextureParameteri(f.TextureID, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(f.TextureID, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(f.TextureID, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TextureParameteri(f.TextureID, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	err := f.LoadFont(filename)
	if err != nil {
		panic(err)
	}
	for _, r := range " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890,./;'#][]" {
		f.DrawRune(r)
	}
	b := &sfnt.Buffer{}
	k, err := f.font.Kern(b, sfnt.GlyphIndex(f.gmap[0].Idx), sfnt.GlyphIndex(f.gmap[1].Idx), f.ppem, 0)
	if err != nil {
		panic(err)
	}
	f.Kern = float32(k) / 64
	return f
}

func (font *Font) LoadFont(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	font.buffer = file
	f, err := sfnt.Parse(font.buffer)
	if err != nil {
		return err
	}
	font.font = f
	font.ppem = (fixed.Int26_6(font.pointSize * font.ppi / 72 * 64)) //fixed.Int26_6(f.UnitsPerEm())
	font.gmap = make(map[rune]Character)
	b := &sfnt.Buffer{}
	m, err := f.Metrics(b, font.ppem, 0)
	if err != nil {
		return err
	}
	font.height = m.Descent.Floor() + m.Ascent.Floor()
	font.VerticalSpacing = float32(m.Height.Ceil())
	gi, err := f.GlyphIndex(b, 'X')
	if err != nil {
		return err
	}
	ink, adv, err := f.GlyphBounds(b, gi, font.ppem, 0)
	if err != nil {
		return err
	}
	minX := min(0, ink.Min.X.Floor())
	maxX := max(adv.Ceil(), ink.Max.X.Ceil())
	minY := min(-m.Ascent.Ceil(), ink.Min.Y.Floor())
	maxY := max(m.Descent.Ceil(), ink.Max.Y.Ceil())

	font.height = maxY - minY
	font.width = maxX - minX
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TextureStorage3D(font.TextureID,
		1,
		gl.R8,
		int32(font.width), int32(font.height),
		int32(font.font.NumGlyphs()))
	return nil
}

func (font *Font) LoadGlyph(tex rune) (sfnt.Segments, uint16, error) {
	b := &sfnt.Buffer{}
	x, err := font.font.GlyphIndex(b, tex)
	if err != nil {
		return nil, 0, err
	}
	if x == 0 {
		return nil, 0, errors.New("No glyph index found for this rune")
	}
	segments, err := font.font.LoadGlyph(b, x, font.ppem, nil)
	if err != nil {
		return nil, 0, err
	}
	return segments, uint16(x), nil
}

// 8-bit gray scale image
func (font *Font) DrawRune(t rune) error {
	segments, gIdx, err := font.LoadGlyph(t)
	if err != nil {
		return err
	}
	b := &sfnt.Buffer{}
	bounds := segments.Bounds()
	advance, err := font.font.GlyphAdvance(b, sfnt.GlyphIndex(gIdx), fixed.Int26_6(font.ppem), 0)
	if err != nil {
		return err
	}
	minX := bounds.Min.X.Floor()
	w, h := font.width, font.height
	var originX, originY float32
	originX = float32(-minX)
	originY = float32(-bounds.Min.Y.Floor())

	r := vector.NewRasterizer(w, h)
	r.DrawOp = draw.Src
	for _, seg := range segments {
		switch seg.Op {
		case sfnt.SegmentOpMoveTo:
			r.MoveTo(
				originX+float32(seg.Args[0].X)/64,
				originY+float32(seg.Args[0].Y)/64,
			)
		case sfnt.SegmentOpLineTo:
			r.LineTo(
				originX+float32(seg.Args[0].X)/64,
				originY+float32(seg.Args[0].Y)/64,
			)
		case sfnt.SegmentOpQuadTo:
			r.QuadTo(
				originX+float32(seg.Args[0].X)/64,
				originY+float32(seg.Args[0].Y)/64,
				originX+float32(seg.Args[1].X)/64,
				originY+float32(seg.Args[1].Y)/64,
			)
		case sfnt.SegmentOpCubeTo:
			r.CubeTo(
				originX+float32(seg.Args[0].X)/64,
				originY+float32(seg.Args[0].Y)/64,
				originX+float32(seg.Args[1].X)/64,
				originY+float32(seg.Args[1].Y)/64,
				originX+float32(seg.Args[2].X)/64,
				originY+float32(seg.Args[2].Y)/64,
			)
		}
	}
	dst := image.NewAlpha(image.Rect(0, 0, w, h))
	r.Draw(dst, dst.Bounds(), image.Opaque, image.Point{})

	const asciiArt = ".++8"
	buf := make([]byte, 0, h*(w+1))
	for y := range h {
		for x := range w {
			a := dst.AlphaAt(x, y).A
			buf = append(buf, asciiArt[a>>6])
		}
		buf = append(buf, '\n')
	}
	os.Stdout.Write(buf)

	if len(dst.Pix) > 0 {
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		gl.TextureSubImage3D(font.TextureID, 0, 0, 0,
			int32(gIdx),
			int32(w),
			int32(h), 1, gl.RED, gl.UNSIGNED_BYTE,
			gl.Ptr(dst.Pix))
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)
	}
	font.gmap[t] = Character{
		Idx:     gIdx,
		Advance: float32(advance.Round()),
		Size: y3d.Vec2{
			X: float32(w),
			Y: float32(h),
		},
	}
	return nil
}

func (font *Font) GetCharacter(r rune) (Character, bool) {
	c, ok := font.gmap[r]
	return c, ok
}

func (font *Font) Destory() {
	gl.DeleteTextures(1, &font.TextureID)
	font.buffer = font.buffer[:0]
}
