package yphy

import (
	"yam/y3d"
	"yam/ywd"
)

type ForceGenerator interface {
	Force(w *ywd.World, e ywd.EntityId, dt float32) y3d.Vec3
}

type DragForce struct {
	CoeffK1 float32
	CoeffK2 float32
}

func NewDragForce(k1, k2 float32) *DragForce {
	return &DragForce{
		CoeffK1: k1,
		CoeffK2: k2,
	}
}

func (d *DragForce) Force(w *ywd.World,
	e ywd.EntityId, dt float32) y3d.Vec3 {
	velocity := w.GetComponent(e, ywd.VelocityComponent).(y3d.Vec3)
	speed := velocity.Length()
	dc := d.CoeffK1*speed + d.CoeffK2*speed*speed
	n := y3d.Normalize(velocity)
	f := y3d.Smul(n, -dc)
	return f
}
