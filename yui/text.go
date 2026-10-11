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
	font            *sfnt.Font
	width           int
	height          int
	ppem            fixed.Int26_6
	gmap            map[rune]Character
}

type Character struct {
	Advance float32
	Min     y3d.Vec2
	Max     y3d.Vec2
	Idx     uint16
}

func NewFont(width, height int, filename string) *Font {
	f := &Font{
		width:  width,
		height: height,
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
	font.ppem = fixed.Int26_6(f.UnitsPerEm())
	font.gmap = make(map[rune]Character)
	b := &sfnt.Buffer{}
	m, err := f.Metrics(b, font.ppem, 0)
	if err != nil {
		return err
	}
	font.VerticalSpacing = float32(m.Height.Ceil())
	gl.TextureStorage3D(font.TextureID,
		1,
		gl.RED,
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
	r := vector.NewRasterizer(font.width, font.height)
	r.DrawOp = draw.Src
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
	var originX, originY float32
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
	dst := image.NewAlpha(image.Rect(0, 0, font.width, font.height))
	r.Draw(dst, dst.Bounds(), image.Opaque, image.Point{})
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TextureSubImage3D(font.TextureID, 0, 0, 0,
		int32(gIdx),
		int32(font.width),
		int32(font.height), 1, gl.RED, gl.UNSIGNED_BYTE,
		gl.Ptr(dst.Pix))
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)

	font.gmap[t] = Character{
		Idx:     gIdx,
		Advance: float32(advance) / 64,
		Min: y3d.Vec2{
			X: (float32(bounds.Min.X) / 64),
			Y: (float32(bounds.Min.Y) / 64),
		},
		Max: y3d.Vec2{
			X: (float32(bounds.Max.X) / 64),
			Y: (float32(bounds.Max.Y) / 64),
		},
	}
	return nil
}

func (font *Font) GetCharacter(r rune) (Character, bool) {
	c, ok := font.gmap[r]
	return c, ok
}
