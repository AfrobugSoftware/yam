package ygame

import (
	"errors"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	ErrorNoSuchSurface = errors.New("no such surface resource")
)

type TextureBundle struct {
	Name     string
	Surfaces []*sdl.Texture
	uses     int
}

type ResourceManager struct {
	SurfaceResource map[string]*TextureBundle
	Renderer        *sdl.Renderer
}

func NewResourceManager(r *sdl.Renderer) *ResourceManager {
	return &ResourceManager{
		Renderer:        r,
		SurfaceResource: map[string]*TextureBundle{},
	}
}

func (r *ResourceManager) GetSurface(name string) (*TextureBundle, error) {
	bundle, exists := r.SurfaceResource[name]
	if !exists {
		return nil, ErrorNoSuchSurface
	}
	bundle.uses++
	return bundle, nil
}

func (r *ResourceManager) ReturnSurface(name string) error {
	bundle, exists := r.SurfaceResource[name]
	if !exists {
		return ErrorNoSuchSurface
	}
	bundle.uses--
	if bundle.uses <= 0 {
		for _, b := range bundle.Surfaces {
			b.Destroy()
		}
		delete(r.SurfaceResource, name)
	}
	return nil
}

func (r *ResourceManager) LoadSurfaceBundle(name string, files []string) error {
	//already exits
	_, ex := r.SurfaceResource[name]
	if ex {
		return nil
	}
	bundle := &TextureBundle{}
	for _, f := range files {
		surface, err := img.Load(f)
		if err != nil {
			//remove any surface already loaded
			for _, b := range bundle.Surfaces {
				b.Destroy()
			}
			return err
		}
		tex, err := r.Renderer.CreateTextureFromSurface(surface)
		if err != nil {
			return err
		}
		bundle.Surfaces = append(bundle.Surfaces, tex)
	}
	r.SurfaceResource[name] = bundle
	return nil
}
