package yrender

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Renderer struct {
	Window   *sdl.Window
	Renderer *sdl.Renderer
	Height   int32
	Width    int32
}

func CreateSDLRenderer(title string, width, height int32) (*Renderer, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, err
	}
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height,
		sdl.WINDOW_OPENGL|sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		return nil, err
	}
	return &Renderer{
		Window:   window,
		Renderer: renderer,
		Width:    width,
		Height:   height,
	}, nil
}

func (re *Renderer) ClearBackgroundColor(r, g, b, a uint8) {
	re.Renderer.SetDrawColor(r, g, b, a)
}

func (r *Renderer) Clear() {
	r.Renderer.Clear()
}

func (r *Renderer) Swap() {
	r.Renderer.Present()
}

func (r *Renderer) Close() {
	if r.Window != nil {
		r.Window.Destroy()
	}
	if r.Renderer != nil {
		r.Renderer.Destroy()
	}
}
