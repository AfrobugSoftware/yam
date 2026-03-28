package yecs

import "yam/y3d"

type Move struct {
	AnglularSpeed float32
	ForwardSpeed  float32
}

type MoveSystem2D struct{}

func (ms *MoveSystem2D) Init()     {}
func (ms *MoveSystem2D) Shutdown() {}

func (ms *MoveSystem2D) Query() []ComponentId {
	return []ComponentId{TransformComponent, MoveComponent}
}

func (ms *MoveSystem2D) Update(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		move := w.GetComponent(e, MoveComponent).(Move)
		transform := w.GetComponent(e, TransformComponent).(Transform)
		if move.AnglularSpeed > y3d.NearZero {
			angle := move.AnglularSpeed * float32(dt)
			inc := y3d.FromAngleAxis(y3d.UNIT_Z, float64(angle))
			transform.Orientation = y3d.ProdQuaternion(inc, transform.Orientation)
			transform.NeedCalculation = true
		}
		if move.ForwardSpeed > y3d.NearZero {
			velocity := y3d.Smul(transform.GetForward(), move.ForwardSpeed)
			transform.Position = y3d.Add(transform.Position, y3d.Smul(velocity, float32(dt)))
			transform.NeedCalculation = true
		}
		if transform.NeedCalculation {
			w.SetComponent(e, TransformComponent, transform)
		}
	}
}
