package ycore

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math/bits"
	"strings"
	"time"
	"yam/y3d"
	"yam/ygl"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	NO_SKINID = -1
)

var (
	EmptySkin = Skin{
		Material: -1,
	}
	EmptyTexture = -1
)

type TextureData struct {
	Handle       uint32
	Size         uint32
	LastAccessed time.Time
	FileOnDisc   string
}

type Skin struct {
	Material int
	Texture  [8]int
}

type SkinManager struct {
	TotalTextureSizeInMemeory uint
	Skins                     map[int]Skin
	Textures                  map[int]TextureData
	Materials                 map[int]ygl.Material
}

func NewSkinManager() *SkinManager {
	return &SkinManager{
		Skins:     make(map[int]Skin),
		Textures:  make(map[int]TextureData),
		Materials: make(map[int]ygl.Material),
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

func generateRandomId() int {
	return int(time.Now().UnixNano())
}

func (s *SkinManager) AddSkin(mat ygl.Material) int {
	skin := Skin{}
	for i := range skin.Texture {
		skin.Texture[i] = EmptyTexture
	}
	id := generateRandomId()
	skinId := generateRandomId()
	s.Materials[id] = mat
	skin.Material = id
	s.Skins[skinId] = skin
	return skinId
}

func (s *SkinManager) FindTextureByFile(filename string) (int, bool) {
	for id, tex := range s.Textures {
		if tex.FileOnDisc == filename {
			return id, true
		}
	}
	return 0, false
}

func (s *SkinManager) AddTexture(skin int, filename string,
	minFilter, maxFilter int32,
	wraps, wrapt int32,
	useMipmap bool) error {
	if _, exists := s.Skins[skin]; !exists {
		return errors.New("invalid skin id")
	}
	if s.Skins[skin].Texture[7] != EmptyTexture {
		return errors.New("skin texture slots are full")
	}
	id, found := s.FindTextureByFile(filename)
	if found {
		sk := s.Skins[skin]
		for slot := range sk.Texture {
			if sk.Texture[slot] == EmptyTexture {
				sk.Texture[slot] = id
				break
			}
		}
		s.Skins[skin] = sk
		return nil
	}
	surface, err := img.Load(filename)
	if err != nil {
		return fmt.Errorf("img.Load(%q): %w", filename, err)
	}
	defer surface.Free()
	converted, err := surface.ConvertFormat(uint32(sdl.PIXELFORMAT_RGBA32), 0)
	if err != nil {
		return fmt.Errorf("ConvertFormat: %w", err)
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
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		w,
		h,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(converted.Pixels()),
	)
	if useMipmap {
		var v float32
		gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY, &v)
		gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAX_ANISOTROPY, v)
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}
	gl.BindTexture(gl.TEXTURE_2D, 0)
	td := TextureData{
		Handle:       texId,
		Size:         uint32(len(converted.Pixels())),
		LastAccessed: time.Now(),
		FileOnDisc:   filename,
	}
	id = generateRandomId()
	s.Textures[id] = td
	sk := s.Skins[skin]
	for slot := range sk.Texture {
		if sk.Texture[slot] == EmptyTexture {
			sk.Texture[slot] = id
			break
		}
	}
	s.Skins[skin] = sk
	return nil
}

func (s *SkinManager) AddSpriteSheet(skin int, filename string, spriteHeight, spriteWidth int,
	minFilter, maxFilter int32,
	wraps, wrapt int32) error {
	if spriteWidth <= 0 || spriteHeight <= 0 {
		return errors.New("sprite dimensions must be positive")
	}
	if _, exists := s.Skins[skin]; !exists {
		return errors.New("invalid skin id")
	}
	if s.Skins[skin].Texture[7] != EmptyTexture {
		return errors.New("skin texture slots are full")
	}
	id, found := s.FindTextureByFile(filename)
	if found {
		sk := s.Skins[skin]
		for slot := range sk.Texture {
			if sk.Texture[slot] == EmptyTexture {
				sk.Texture[slot] = id
				break
			}
		}
		s.Skins[skin] = sk
		return nil
	}
	surface, err := img.Load(filename)
	if err != nil {
		return fmt.Errorf("img.Load(%q): %w", filename, err)
	}
	defer surface.Free()
	converted, err := surface.ConvertFormat(uint32(sdl.PIXELFORMAT_RGBA32), 0)
	if err != nil {
		return fmt.Errorf("ConvertFormat: %w", err)
	}
	defer converted.Free()
	if converted.MustLock() {
		if err := converted.Lock(); err != nil {
			return fmt.Errorf("surface lock: %w", err)
		}
		defer converted.Unlock()
	}
	w, h := int32(converted.W), int32(converted.H)
	var texId uint32
	cols := w / int32(spriteWidth)
	rows := h / int32(spriteHeight)
	if cols == 0 || rows == 0 {
		return errors.New("sprite larger than sheet")
	}

	pix := converted.Pixels()
	pitch := int(converted.Pitch)
	bpp := converted.BytesPerPixel()
	if int(pitch)%bpp != 0 {
		return errors.New("surface pitch is not a multiple of pixel size")
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
		return fmt.Errorf("GL error 0x%x uploading %q", e, filename)
	}
	td := TextureData{
		Handle:       texId,
		Size:         uint32(len(converted.Pixels())),
		LastAccessed: time.Now(),
		FileOnDisc:   filename,
	}
	id = generateRandomId()
	s.Textures[id] = td
	sk := s.Skins[skin]
	for slot := range sk.Texture {
		if sk.Texture[slot] == EmptyTexture {
			sk.Texture[slot] = id
			break
		}
	}
	s.Skins[skin] = sk
	return nil
}

func (s *SkinManager) GetTexture(skin int, slot int) (uint32, error) {
	if _, exists := s.Skins[skin]; !exists {
		return 0, errors.New("invalid skin id")
	}
	sk := s.Skins[skin]
	tex := s.Textures[sk.Texture[slot]]
	return tex.Handle, nil
}

func (s *SkinManager) GetMaterial(skin int) (ygl.Material, error) {
	if _, exists := s.Skins[skin]; !exists {
		return ygl.Material{}, errors.New("invalid skin id")
	}
	sk := s.Skins[skin]
	return s.Materials[sk.Material], nil
}

func (s *SkinManager) GetSkin(skin int) (Skin, error) {
	if _, exists := s.Skins[skin]; !exists {
		return Skin{}, errors.New("invalid skin id")
	}
	return s.Skins[skin], nil
}

func (s *SkinManager) RemoveTexture(skin int, slot int) error {
	if _, exists := s.Skins[skin]; !exists {
		return errors.New("invalid skin id")
	}
	sk := s.Skins[skin]
	if sk.Texture[slot] == EmptyTexture {
		return errors.New("texture slot is empty")
	}
	handle := s.Textures[sk.Texture[slot]].Handle
	gl.DeleteTextures(1, &handle)
	s.TotalTextureSizeInMemeory -= uint(s.Textures[sk.Texture[slot]].Size)
	delete(s.Textures, sk.Texture[slot])
	sk.Texture[slot] = EmptyTexture
	s.Skins[skin] = sk
	return nil
}

func (s *SkinManager) RemoveSkin(skin int) {
	if _, exists := s.Skins[skin]; !exists {
		return
	}
	for _, texId := range s.Skins[skin].Texture {
		if texId != EmptyTexture {
			handle := s.Textures[texId].Handle
			gl.DeleteTextures(1, &handle)
			s.TotalTextureSizeInMemeory -= uint(s.Textures[texId].Size)
			delete(s.Textures, texId)
		}
	}
	delete(s.Skins, skin)
}

func (s *SkinManager) ConvertHeightMapToNormalMap(hmap *image.RGBA) (*image.RGBA, error) {
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

func (s *SkinManager) Destroy() {
	clear(s.Materials)
	for _, t := range s.Textures {
		gl.DeleteTextures(1, &t.Handle)
	}
	clear(s.Textures)
	clear(s.Skins)
}

func (s *SkinManager) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Total texture size in memory: %d\n", s.TotalTextureSizeInMemeory)
	fmt.Fprintf(&b, "Skins: %v\n", s.Skins)
	return b.String()
}
