package ygame

import "yam/y3d"

type MoveComponent struct{}

func (m MoveComponent) UpdateComponent(dt float64, s *Sprite) {
	s.Pos = y3d.Add(s.Pos, y3d.Smul(s.Vel, dt))
}
