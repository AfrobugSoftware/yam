package ymanager

import (
	"errors"
	"image"
	"os"
	"time"
	"yam/yecs"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	SLOT_1 int = 0
	SLOT_2 int = 0
	SLOT_3 int = 0
	SLOT_4 int = 0
	SLOT_5 int = 0
	SLOT_6 int = 0
	SLOT_7 int = 0
	SLOT_8 int = 0
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
	Skins                     []Skin
	Textures                  []TextureData
	Materials                 []yecs.Material
}

func NewSkinManager() *SkinManager {
	return &SkinManager{
		Skins:     make([]Skin, 0),
		Textures:  make([]TextureData, 0),
		Materials: make([]yecs.Material, 0),
	}
}

func (s *SkinManager) AddTexture(skin int, slot int, filename string, minFilter, maxFilter int32, useMipmap bool) {
	if skin >= len(s.Skins) {
		panic("invalid skin id")
	}
	fileInfo, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}
	texSize := fileInfo.Size()
	s.TotalTextureSizeInMemeory += uint(texSize)
	//need to check if the size if bigger than the budget

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	var texId uint32
	gl.GenTextures(1, &texId)
	gl.BindTexture(gl.TEXTURE_2D, texId)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
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
	s.Textures = append(s.Textures, td)
	sk := &s.Skins[skin]
	sk.Texture[slot] = len(s.Textures) - 1
}

func (s *SkinManager) GetTexture(skin int, slot int) (uint32, error) {
	if skin >= len(s.Skins) {
		return 0, errors.New("invalid skin id")
	}
	sk := &s.Skins[skin]
	tex := s.Textures[sk.Texture[slot]]
	return tex.Handle, nil
}
