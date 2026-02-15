package yrender

import "yam/y3d"

type Sprite struct {
	Pos           y3d.Vec3
	Vel           y3d.Vec3
	Width, Height int32
}

func (s *Sprite) ProcessInput(dt float64) {

}

func (s *Sprite) Update(dt float64) {
	s.Pos = y3d.Add(s.Pos, y3d.Smul(s.Vel, dt))
}

func (s *Sprite) Draw(dt float64) {

}
