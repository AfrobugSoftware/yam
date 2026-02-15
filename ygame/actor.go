package ygame

import (
	"yam/y3d"

	"github.com/veandco/go-sdl2/sdl"
)

type Actor interface {
	Update(dt float64)
	Draw(sr *sdl.Renderer)
	ProcessInput()
	GetBox() y3d.AABB
	GetType() int
}
