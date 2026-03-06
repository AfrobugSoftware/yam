package ygame

import "yam/y3d"

type MoveComponent struct {
	ForwardSpeed float64
	AngularSpeed float64
	Owner        Actor
}

func (m *MoveComponent) UpdateComponent(dt float64, s *Sprite) {
	s.Angle += m.AngularSpeed * dt
	v := y3d.Smul(s.GetForward(), m.ForwardSpeed)
	s.Pos = y3d.Add(s.Pos, y3d.Smul(v, dt))
}

func (m *MoveComponent) ProcessInput(keys []uint8) {

}

type InputComponent struct {
	MoveComponent
	OtherProcess         func(keys []uint8)
	MaxForwardSpeed      float64
	MaxAngularSpeed      float64
	ForwardKey           int
	BackwardKey          int
	ClockwiseRotationKey int
	CounterClockwiseKey  int
}

func (i *InputComponent) ProcessInput(keys []uint8) {
	var forwardSpeed, anglurSpeed float64
	if keys[i.ForwardKey] != 0 {
		forwardSpeed += i.MaxForwardSpeed
	}
	if keys[i.BackwardKey] != 0 {
		forwardSpeed -= i.MaxForwardSpeed
	}
	if i.ClockwiseRotationKey != -1 && keys[i.ClockwiseRotationKey] != 0 {
		anglurSpeed += i.MaxAngularSpeed
	}
	if i.CounterClockwiseKey != -1 && keys[i.CounterClockwiseKey] != 0 {
		anglurSpeed -= i.MaxAngularSpeed
	}
	i.MoveComponent.ForwardSpeed = forwardSpeed
	i.MoveComponent.AngularSpeed = anglurSpeed
	if i.OtherProcess != nil {
		i.OtherProcess(keys)
	}
}
