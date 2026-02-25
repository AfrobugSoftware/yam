package ygame

import (
	"yam/yutil"

	"github.com/veandco/go-sdl2/sdl"
)

type AnimSprite struct {
	Sprite
	FPS         float64
	CurFrame    float64
	Frames      *sdl.Texture
	NumFrames   int
	FrameHeight int32
	FrameWidth  int32
}

func (an *AnimSprite) SelectSheet(indx int) {
	an.Frames = an.Texture.Surfaces[indx]
}

func (an *AnimSprite) Update(dt float64) {
	an.Sprite.Update(dt)
	an.CurFrame += an.FPS * dt
	for an.CurFrame >= float64(an.NumFrames) {
		an.CurFrame -= float64(an.NumFrames)
	}
}

func (s *AnimSprite) Draw(renderer *sdl.Renderer) {
	srcRect := &sdl.Rect{
		X: int32(s.CurFrame) * s.FrameWidth,
		Y: 0,
		W: s.FrameWidth,
		H: s.FrameHeight,
	}
	dstRect := &sdl.Rect{
		X: int32(s.Pos.X) - int32(s.FrameWidth/2),
		Y: int32(s.Pos.Y) - int32(s.FrameHeight/2),
		W: s.FrameWidth,
		H: s.FrameHeight,
	}
	renderer.CopyEx(s.Frames, srcRect, dstRect, -yutil.ToDegree(s.Angle), nil, sdl.FLIP_NONE)
}
