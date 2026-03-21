package ygame

import "yam/yecs"

type Renderer struct {
	World      *yecs.World
	MainCamera *Camera
}

func NewRenderer(width, height int, w *yecs.World) *Renderer {
	return nil
}

func (render *Renderer) Draw() {

}
