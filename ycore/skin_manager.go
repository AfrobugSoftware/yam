package ycore

import (
	"errors"
	"fmt"
	"image"
	"os"
	"strings"
	"time"
	"yam/ygl"

	"github.com/go-gl/gl/v4.3-core/gl"
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
	Alpha    bool
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
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	texSize := len(rgba.Pix)
	s.TotalTextureSizeInMemeory += uint(texSize)
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
		int32(bounds.Max.X),
		int32(bounds.Max.Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(&rgba.Pix[0]),
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
		Size:         uint32(texSize),
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

func (s *SkinManager) ConvertHeightMapToNormalMap(hmap *image.Image) (*image.Image, error) {
	if hmap == nil {
		return nil, errors.New("no height map given")
	}
	return nil, nil
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
