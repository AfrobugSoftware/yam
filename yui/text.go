package yui

import (
	"errors"
	"image"
	"image/draw"
	"os"
	"unicode/utf8"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

const atlasPadding = 1
const defaultCharset = " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890,!./;'#][]"

var ErrAtlasFull = errors.New("font atlas has no space left for glyph")

type FontInfo struct {
	UMin, VMin float32
	UMax, VMax float32
}

type FontAtlas struct {
	ID     uint32
	Width  int
	Height int
	UVs    map[rune]FontInfo
}

type Font struct {
	Atlas           FontAtlas
	Ascent          float32
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

	// shelf packer state
	atlasX, atlasY int
	atlasRowH      int
}

type Character struct {
	Advance float32
	Size    y3d.Vec2
	Bearing y3d.Vec2
	Idx     uint16
}

func NewFont(pointSize, ppi float32, filename string) *Font {
	f := &Font{
		pointSize: pointSize,
		ppi:       ppi,
	}

	err := f.LoadFont(filename)
	if err != nil {
		panic(err)
	}
	f.initAtlas(utf8.RuneCountInString(defaultCharset))
	for _, r := range defaultCharset {
		f.DrawRune(r)
	}
	b := &sfnt.Buffer{}
	k, err := f.font.Kern(b, sfnt.GlyphIndex(f.gmap[0].Idx), sfnt.GlyphIndex(f.gmap[1].Idx), f.ppem, 0)
	if err != nil {
		panic(err)
	}
	f.Kern = float32(k.Ceil())
	return f
}

func (font *Font) initAtlas(glyphCount int) {
	cellW := font.width + atlasPadding
	cellH := font.height + atlasPadding

	size := 64
	for ((size-atlasPadding)/cellW)*((size-atlasPadding)/cellH) < glyphCount {
		size *= 2
	}

	font.Atlas = FontAtlas{
		Width:  size,
		Height: size,
		UVs:    make(map[rune]FontInfo),
	}

	gl.CreateTextures(gl.TEXTURE_2D, 1, &font.Atlas.ID)
	gl.TextureParameteri(font.Atlas.ID, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(font.Atlas.ID, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(font.Atlas.ID, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TextureParameteri(font.Atlas.ID, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TextureStorage2D(font.Atlas.ID, 1, gl.R8, int32(size), int32(size))

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	zero := make([]byte, size*size)
	gl.TextureSubImage2D(font.Atlas.ID, 0, 0, 0, int32(size), int32(size),
		gl.RED, gl.UNSIGNED_BYTE, gl.Ptr(zero))

	font.atlasX = atlasPadding
	font.atlasY = atlasPadding
	font.atlasRowH = 0
}

func (font *Font) allocate(w, h int) (int, int, error) {
	if font.atlasX+w+atlasPadding > font.Atlas.Width {
		// start a new shelf
		font.atlasX = atlasPadding
		font.atlasY += font.atlasRowH + atlasPadding
		font.atlasRowH = 0
	}
	if font.atlasX+w+atlasPadding > font.Atlas.Width ||
		font.atlasY+h+atlasPadding > font.Atlas.Height {
		return 0, 0, ErrAtlasFull
	}
	x, y := font.atlasX, font.atlasY
	font.atlasX += w + atlasPadding
	font.atlasRowH = max(font.atlasRowH, h)
	return x, y, nil
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
	font.VerticalSpacing = float32(m.Height.Ceil())
	font.Ascent = float32(m.Ascent.Ceil())
	gi, err := f.GlyphIndex(b, 'W')
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

func (font *Font) DrawRune(t rune) error {
	if _, ok := font.gmap[t]; ok {
		return nil
	}
	segments, gIdx, err := font.LoadGlyph(t)
	if err != nil {
		return err
	}
	b := &sfnt.Buffer{}
	advance, err := font.font.GlyphAdvance(b, sfnt.GlyphIndex(gIdx), font.ppem, 0)
	if err != nil {
		return err
	}
	bounds := segments.Bounds()
	minX, minY := bounds.Min.X.Floor(), bounds.Min.Y.Floor()
	maxX, maxY := bounds.Max.X.Ceil(), bounds.Max.Y.Ceil()
	w, h := maxX-minX, maxY-minY

	var info FontInfo
	if w > 0 && h > 0 { // whitespace has no ink: no atlas space, zero UVs
		ax, ay, err := font.allocate(w, h)
		if err != nil {
			return err
		}

		originX := float32(-minX)
		originY := float32(-minY)

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
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		gl.TextureSubImage2D(font.Atlas.ID, 0,
			int32(ax), int32(ay),
			int32(w), int32(h),
			gl.RED, gl.UNSIGNED_BYTE,
			gl.Ptr(dst.Pix))

		aw, ah := float32(font.Atlas.Width), float32(font.Atlas.Height)
		info = FontInfo{
			UMin: float32(ax) / aw,
			VMin: float32(ay) / ah,
			UMax: float32(w) / aw,
			VMax: float32(h) / ah,
		}
	}

	font.Atlas.UVs[t] = info
	font.gmap[t] = Character{
		Idx:     gIdx,
		Advance: float32(advance.Round()),
		Size: y3d.Vec2{
			X: float32(w),
			Y: float32(h),
		},
		Bearing: y3d.Vec2{
			X: float32(minX),
			Y: float32(minY),
		},
	}
	return nil
}

func (font *Font) GetCharacter(r rune) (Character, bool) {
	c, ok := font.gmap[r]
	return c, ok
}

func (font *Font) GetUV(r rune) (FontInfo, bool) {
	uv, ok := font.Atlas.UVs[r]
	return uv, ok
}

func (font *Font) Destory() {
	gl.DeleteTextures(1, &font.Atlas.ID)
	font.buffer = font.buffer[:0]
}
