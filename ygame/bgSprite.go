package ygame

import (
	"yam/y3d"

	"github.com/veandco/go-sdl2/sdl"
)

type BGTexture struct {
	Texture *sdl.Texture
	Offset  y3d.Vec3
}

type BGSprite struct {
	Sprite
	ScrollSpeed float64
	Backgrounds []BGSprite
}

func (bg *BGSprite) Update(dt float64) {

}

func (bg *BGSprite) Draw(renderer *sdl.Renderer) {

}
