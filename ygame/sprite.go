package ygame

import (
	"yam/y3d"
	"yam/yutil"

	"github.com/veandco/go-sdl2/sdl"
)

type Sprite struct {
	Pos           y3d.Vec3
	Vel           y3d.Vec3
	Color         [4]uint8
	Box           y3d.AABB
	Scale         float64
	Angle         float64 // in radians
	Width, Height int32
	State         int
	Texture       *TextureBundle
	TexIndx       int //what texture in the bundle
	Components    []Component
	StateMachine  *StateMachine
}

const (
	INVALID_SPRITE = -1
)

func (s *Sprite) GetType() int {
	return INVALID_SPRITE
}

func (s *Sprite) ProcessInput() {
	if s.State == ACTIVE {
		g := GetGame()
		for _, c := range s.Components {
			c.ProcessInput(g.KeyState)
		}
	}
}

func (s *Sprite) GenerateAABB() {
	s.Box = y3d.AABB{
		MinX: s.Pos.X - float64(s.Width)/2.0,
		MinY: s.Pos.Y - float64(s.Height)/2.0,
		MaxX: s.Pos.X + float64(s.Width)/2.0,
		MaxY: s.Pos.Y + float64(s.Height)/2.0,
	}
}

func (s *Sprite) GetState() int {
	return s.State
}

func (s *Sprite) GetBox() y3d.AABB {
	return s.Box
}

func (s *Sprite) Update(dt float64) {
	for _, comp := range s.Components {
		comp.UpdateComponent(dt, s)
	}
	if s.StateMachine != nil {
		s.StateMachine.Update(dt)
	}
	s.GenerateAABB()
}

func (s *Sprite) GetForward() y3d.Vec3 {
	return y3d.GetForward2D(s.Angle)
}

func (s *Sprite) Draw(sr *sdl.Renderer) {
	rect := &sdl.Rect{
		X: int32(s.Pos.X) - int32(s.Width/2),
		Y: int32(s.Pos.Y) - int32(s.Height/2),
		W: s.Width,
		H: s.Height,
	}
	if s.Texture != nil && s.TexIndx != -1 {
		tex := s.Texture.Surfaces[s.TexIndx]
		sr.CopyEx(tex, nil, rect, -yutil.ToDegree(s.Angle), nil, sdl.FLIP_NONE)
	} else {
		sr.SetDrawColor(
			s.Color[0],
			s.Color[1],
			s.Color[2],
			s.Color[3],
		)
		sr.FillRect(rect)
	}
}

func (s *Sprite) GetRect() sdl.Rect {
	return sdl.Rect{
		X: int32(s.Pos.X) - int32(s.Width/2),
		Y: int32(s.Pos.Y) - int32(s.Height/2),
		W: s.Width,
		H: s.Height,
	}
}
