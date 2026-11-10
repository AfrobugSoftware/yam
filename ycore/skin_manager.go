package ycore

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/bits"
	"math/rand"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type TextureManager struct {
	Textures map[string]uint32
}

func NewSkinManager() *TextureManager {
	return &TextureManager{
		Textures: make(map[string]uint32),
	}
}

func mipLevels(minFilter, w, h int32) int32 {
	switch minFilter {
	case gl.NEAREST_MIPMAP_NEAREST, gl.LINEAR_MIPMAP_NEAREST,
		gl.NEAREST_MIPMAP_LINEAR, gl.LINEAR_MIPMAP_LINEAR:
		return int32(bits.Len32(uint32(max(w, h))))
	}
	return 1
}

func GenerateRandomId() int {
	return rand.Intn(math.MaxInt-1) + 1 //never generates 0
}

func (s *TextureManager) AddTexture(filename string,
	minFilter, maxFilter int32,
	wraps, wrapt int32) (uint32, error) {

	id, found := s.Textures[filename]
	if found {
		return id, nil
	}
	surface, err := img.Load(filename)
	if err != nil {
		return 0, fmt.Errorf("img.Load(%q): %w", filename, err)
	}
	defer surface.Free()
	converted, err := surface.ConvertFormat(uint32(sdl.PIXELFORMAT_RGBA32), 0)
	if err != nil {
		return 0, fmt.Errorf("ConvertFormat: %w", err)
	}
	defer converted.Free()
	w, h := int32(converted.W), int32(converted.H)
	var texId uint32
	gl.GenTextures(1, &texId)
	gl.BindTexture(gl.TEXTURE_2D, texId)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, wraps)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, wrapt)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, minFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, maxFilter)
	levels := mipLevels(minFilter, w, h)
	gl.TextureStorage2D(texId, levels, gl.RGBA8, w, h)
	gl.TextureSubImage2D(texId, 0, 0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(converted.Pixels()))
	if levels > 1 {
		var v float32
		gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY, &v)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAX_ANISOTROPY, v)
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}
	gl.BindTexture(gl.TEXTURE_2D, 0)
	s.Textures[filename] = texId
	return texId, nil
}

func (s *TextureManager) AddCubeMap(filename [6]string) (uint32, error) {
	id, found := s.Textures[filename[0]]
	if found {
		return id, nil
	}
	var texId uint32
	gl.CreateTextures(gl.TEXTURE_CUBE_MAP, 1, &texId)
	gl.TextureParameteri(texId, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(texId, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(texId, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TextureParameteri(texId, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	for i, s := range filename {
		surface, err := img.Load(s)
		if err != nil {
			return 0, fmt.Errorf("img.Load(%q): %w", s, err)
		}
		converted, err := surface.ConvertFormat(uint32(sdl.PIXELFORMAT_RGBA32), 0)
		surface.Free()
		if err != nil {
			return 0, fmt.Errorf("ConvertFormat: %w", err)
		}
		w, h := int32(converted.W), int32(converted.H)
		if w != h {
			converted.Free()
			return 0, fmt.Errorf("cube face %q is %dx%d, must be square", s, w, h)
		}
		if i == 0 {
			gl.TextureStorage2D(texId, 1, gl.RGBA8, w, h)
		}
		if converted.MustLock() {
			if err := converted.Lock(); err != nil {
				converted.Free()
				converted.Unlock()
				return 0, fmt.Errorf("Lock(%q): %w", s, err)
			}
			converted.Unlock()
		}
		gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
		gl.PixelStorei(gl.UNPACK_ROW_LENGTH, converted.Pitch/4)
		gl.TextureSubImage3D(texId, 0, 0, 0, int32(i), w, h, 1,
			gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(converted.Pixels()))
		gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
		converted.Free()
	}
	s.Textures[filename[0]] = texId
	return texId, nil
}

func (s *TextureManager) AddSpriteSheet(filename string, spriteHeight, spriteWidth int,
	minFilter, maxFilter int32,
	wraps, wrapt int32) (uint32, error) {
	if spriteWidth <= 0 || spriteHeight <= 0 {
		return 0, errors.New("sprite dimensions must be positive")
	}
	id, found := s.Textures[filename]
	if found {
		return id, nil
	}
	surface, err := img.Load(filename)
	if err != nil {
		return 0, fmt.Errorf("img.Load(%q): %w", filename, err)
	}
	defer surface.Free()
	converted, err := surface.ConvertFormat(uint32(sdl.PIXELFORMAT_RGBA32), 0)
	if err != nil {
		return 0, fmt.Errorf("ConvertFormat: %w", err)
	}
	defer converted.Free()
	if converted.MustLock() {
		if err := converted.Lock(); err != nil {
			return 0, fmt.Errorf("surface lock: %w", err)
		}
		defer converted.Unlock()
	}
	w, h := int32(converted.W), int32(converted.H)
	var texId uint32
	cols := w / int32(spriteWidth)
	rows := h / int32(spriteHeight)
	if cols == 0 || rows == 0 {
		return 0, errors.New("sprite larger than sheet")
	}

	pix := converted.Pixels()
	pitch := int(converted.Pitch)
	bpp := converted.BytesPerPixel()
	if int(pitch)%bpp != 0 {
		return 0, errors.New("surface pitch is not a multiple of pixel size")
	}
	sw, sh := int32(spriteWidth), int32(spriteHeight)
	layers := int32(cols * rows)
	levels := mipLevels(minFilter, sw, sh)
	gl.CreateTextures(gl.TEXTURE_2D_ARRAY, 1, &texId)
	gl.TextureParameteri(texId, gl.TEXTURE_WRAP_S, wraps)
	gl.TextureParameteri(texId, gl.TEXTURE_WRAP_T, wrapt)
	gl.TextureParameteri(texId, gl.TEXTURE_MIN_FILTER, minFilter)
	gl.TextureParameteri(texId, gl.TEXTURE_MAG_FILTER, maxFilter)
	gl.TextureStorage3D(texId, levels, gl.RGBA8, sw, sh, layers)

	gl.BindBuffer(gl.PIXEL_UNPACK_BUFFER, 0)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, int32(pitch/bpp))
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	for row := 0; row < int(rows); row++ {
		for col := 0; col < int(cols); col++ {
			offset := row*spriteHeight*pitch + col*spriteWidth*bpp
			layer := int32(row*int(cols) + col)
			gl.TextureSubImage3D(texId, 0,
				0, 0, layer,
				sw, sh, 1,
				gl.RGBA, gl.UNSIGNED_BYTE,
				gl.Ptr(pix[offset:]))
		}
	}
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 4)

	if levels > 1 {
		gl.GenerateTextureMipmap(texId)
	}
	if e := gl.GetError(); e != gl.NO_ERROR {
		gl.DeleteTextures(1, &texId)
		return 0, fmt.Errorf("GL error 0x%x uploading %q", e, filename)
	}
	s.Textures[filename] = texId
	return texId, nil
}

func (s *TextureManager) ConvertHeightMapToNormalMap(hmap *image.RGBA) (*image.RGBA, error) {
	if hmap == nil {
		return nil, errors.New("no height map given")
	}
	bounds := hmap.Bounds()
	ret := image.NewRGBA(hmap.Rect)
	for i := range bounds.Max.Y {
		for j := range bounds.Max.X {
			var c00, c10, c01 color.RGBA
			c00 = hmap.RGBAAt(j, i)
			if j+1 >= bounds.Max.X {
				c10 = hmap.RGBAAt(0, i)
			} else {
				c10 = hmap.RGBAAt(j+1, i)
			}
			if i+i >= bounds.Max.Y {
				c01 = hmap.RGBAAt(j, 0)
			} else {
				c01 = hmap.RGBAAt(j, i+1)
			}
			h00 := float32(c00.R / 255.0)
			h10 := float32(c10.R / 255.0)
			h01 := float32(c01.R / 255.0)

			p00 := y3d.Vec3{
				X: float32(j),
				Y: float32(i),
				Z: h00,
			}
			p10 := y3d.Vec3{
				X: float32(j) + 1.0,
				Y: float32(i),
				Z: h10,
			}

			p01 := y3d.Vec3{
				X: float32(j),
				Y: float32(i) + 1.0,
				Z: h01,
			}
			v1 := y3d.Sub(p10, p00)
			v0 := y3d.Sub(p01, p00)
			n := y3d.Normalize(y3d.Cross(v1, v0))
			nc := color.RGBA{
				R: uint8(127.0*n.X + 128.0),
				G: uint8(127.0*n.Y + 128.0),
				B: uint8(127.0*n.Z + 128.0),
				A: uint8(255.0 * h00),
			}
			ret.Set(j, i, nc)
		}
	}
	return ret, nil
}

func (s *TextureManager) Destroy() {
	for _, t := range s.Textures {
		gl.DeleteTextures(1, &t)
	}
	clear(s.Textures)
}
