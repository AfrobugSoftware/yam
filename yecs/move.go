package yecs

import (
	"math"
	"runtime"
	"sync"
	"yam/y3d"
)

type Move struct {
	AnglularSpeed float32
	ForwardSpeed  float32
	StrafeSpeed   float32
	VerticalSpeed float32

	MaxAngularSpeed float32
}

type MoveSystem struct {
	Wg sync.WaitGroup
}

func (ms *MoveSystem) Init()     {}
func (ms *MoveSystem) Shutdown() {}

func (ms *MoveSystem) Query() []ComponentId {
	return []ComponentId{TransformComponent, MoveComponent}
}

func (ms *MoveSystem) Run(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		recal := false
		move := w.GetComponent(e, MoveComponent).(Move)
		transform := w.GetComponent(e, TransformComponent).(Transform)
		if math.Abs(float64(move.AnglularSpeed)) > y3d.NearZero {
			angle := move.AnglularSpeed * float32(dt)
			angle = float32(y3d.ToRadians(float64(angle)))
			inc := y3d.FromAngleAxis(transform.GetForward(), float64(angle))
			transform.Rotation = y3d.ProdQuaternion(inc, transform.Rotation)
			recal = true
		}
		if math.Abs(float64(move.ForwardSpeed)) > y3d.NearZero {
			velocity := y3d.Smul(transform.GetForward(), move.ForwardSpeed)
			transform.Position = y3d.Add(transform.Position, y3d.Smul(velocity, float32(dt)))
			recal = true
		}
		if math.Abs(float64(move.StrafeSpeed)) > y3d.NearZero {
			velocity := y3d.Smul(transform.GetRight(), move.StrafeSpeed)
			transform.Position = y3d.Add(transform.Position, y3d.Smul(velocity, float32(dt)))
			recal = true
		}
		if math.Abs(float64(move.VerticalSpeed)) > y3d.NearZero {
			velocity := y3d.Smul(transform.GetUp(), move.VerticalSpeed)
			transform.Position = y3d.Add(transform.Position, y3d.Smul(velocity, float32(dt)))
			recal = true
		}
		if recal {
			(&transform).Recalulate()
			if w.HasComponent(e, BoxComponent) {
				box := w.GetComponent(e, BoxComponent).(Box)
				box.World = transform.TransFormAABB(box.Local)
				w.SetComponent(e, BoxComponent, box)
			}
			w.SetComponent(e, TransformComponent, transform)
		}
	}
	ms.Wg.Done()
}

func (ms *MoveSystem) Update(w *World, dt float64, entites []EntityId) {
	cpu := runtime.NumCPU()
	brk := len(entites) / cpu
	if brk < cpu {
		ms.Wg.Add(1)
		go ms.Run(w, dt, entites)
	} else {
		//does not work with odd entities
		for i := range cpu {
			ms.Wg.Add(1)
			offset := i * brk
			go ms.Run(w, dt, entites[offset:offset+brk])
		}
	}
	ms.Wg.Wait()
}

type Navigator struct {
	Move
	Positions []y3d.Vec3
	CurPos    int
}

type NavigatorMoveSystem struct{}

func (ns *NavigatorMoveSystem) Init()     {}
func (ns *NavigatorMoveSystem) Shutdown() {}
func (ns *NavigatorMoveSystem) Update(w *World, dt float64, entires []EntityId) {

}
func (ns *NavigatorMoveSystem) Query() []ComponentId {
	return []ComponentId{TransformComponent, NavigatorComponent}
}
