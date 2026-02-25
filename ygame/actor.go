package ygame

import (
	"yam/y3d"

	"github.com/veandco/go-sdl2/sdl"
)

// state
const (
	ACTIVE = iota
	PAUSED
	DEAD
)

// the problem is that, interfaces describes behaviour not data
// so what are actors, store of data or systems of behaviour??
type Actor interface {
	Update(dt float64)
	Draw(sr *sdl.Renderer)
	ProcessInput()
	GetBox() y3d.AABB
	GetType() int
	GetState() int
}

type Component interface {
	UpdateComponent(dt float64, s *Sprite)
}
