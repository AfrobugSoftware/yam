package yam

import (
	"yam/y3d"
	"yam/ygame"
	"yam/yrender"

	"github.com/veandco/go-sdl2/sdl"
)

type Paddle struct {
	yrender.Sprite
}

type Wall struct {
	yrender.Sprite
}

func NewPaddle(pos y3d.Vec3) *Paddle {
	p := &Paddle{}
	p.Width = 5
	p.Height = 10
	p.Pos = pos
	return p
}

func NewWall() *Wall {
	return &Wall{}
}

func (p *Paddle) ProcessInput() {
	g := ygame.GetGame()
	p.Vel.Y = 0
	if g.KeyState[sdl.SCANCODE_W] != 0 {
		p.Vel.Y -= 1
	}
	if g.KeyState[sdl.SCANCODE_S] != 0 {
		p.Vel.Y += 1
	}
}
