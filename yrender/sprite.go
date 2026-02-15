package yrender

import (
	"yam/y3d"

	"github.com/veandco/go-sdl2/sdl"
)

type Sprite struct {
	Pos           y3d.Vec3
	Vel           y3d.Vec3
	Color         [4]uint8
	Box           y3d.AABB
	Width, Height int32
}

const (
	INVALID_SPRITE = -1
)

func (s *Sprite) GetType() int {
	return INVALID_SPRITE
}

func (s *Sprite) ProcessInput() {

}

func (s *Sprite) GenerateAABB() {
	s.Box = y3d.AABB{
		MinX: float64(int32(s.Pos.X) - int32(s.Width/2)),
		MinY: float64(int32(s.Pos.Y) - int32(s.Height/2)),
		MaxX: float64(int32(s.Pos.X)-int32(s.Width/2)) + float64(s.Width),
		MaxY: float64(int32(s.Pos.Y)-int32(s.Height/2)) + float64(s.Height),
	}
}

func (s *Sprite) GetBox() y3d.AABB {
	return s.Box
}

func (s *Sprite) Update(dt float64) {
	s.Pos = y3d.Add(s.Pos, y3d.Smul(s.Vel, dt))
	s.GenerateAABB()
}

func (s *Sprite) Draw(sr *sdl.Renderer) {
	rect := &sdl.Rect{
		X: int32(s.Pos.X) - int32(s.Width/2),
		Y: int32(s.Pos.Y) - int32(s.Height/2),
		W: s.Width,
		H: s.Height,
	}
	sr.SetDrawColor(
		s.Color[0],
		s.Color[1],
		s.Color[2],
		s.Color[3],
	)
	sr.FillRect(rect)
}
