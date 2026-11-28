package yphy

import (
	"yam/y3d"
	"yam/ywd"
)

type ForceGenerator interface {
	Force(w *ywd.World, e ywd.EntityId, dt float32)
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
	e ywd.EntityId, dt float32) {
	velocity := w.GetComponent(e, ywd.VelocityComponent).(y3d.Vec3)
	force := w.GetComponent(e, ywd.ForceComponent).(y3d.Vec3)
	speed := velocity.Length()
	dc := d.CoeffK1*speed + d.CoeffK2*speed*speed
	n := y3d.Normalize(velocity)
	f := y3d.Smul(n, -dc)
	force = y3d.Add(force, f)
	w.SetComponent(e, ywd.ForceComponent, force)
}

func AddForceAtPoint(w *ywd.World,
	e ywd.EntityId, force, point y3d.Vec3) {
	f := w.GetComponent(e, ywd.ForceComponent).(y3d.Vec3)
	t := w.GetComponent(e, ywd.TorqueComponent).(y3d.Vec3)
	position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)

	point = y3d.Sub(point, position)
	torque := y3d.Cross(point, force)
	f = y3d.Add(f, force)
	t = y3d.Add(t, torque)

	w.SetComponent(e, ywd.ForceComponent, f)
	w.SetComponent(e, ywd.TorqueComponent, t)
}

func AddTorque(w *ywd.World,
	e ywd.EntityId, torque y3d.Vec3) {
	t := w.GetComponent(e, ywd.TorqueComponent).(y3d.Vec3)
	t = y3d.Add(t, torque)
	w.SetComponent(e, ywd.TorqueComponent, t)
}

func TransformPointToWorldSpace(w *ywd.World, e ywd.EntityId, point y3d.Vec3) y3d.Vec3 {
	position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
	orientation := w.GetComponent(e, ywd.OrientationComponent).(y3d.Quaternion)
	scale := w.GetComponent(e, ywd.ScaleComponent).(y3d.Vec3)
	transform := y3d.TRS(position,
		orientation,
		scale)
	t := y3d.Vec4{X: point.X, Y: point.Y, Z: point.Z, W: 1.0}
	r := transform.MulVec4(t)
	return y3d.Vec3{X: r.X, Y: r.Y, Z: r.Z}
}
