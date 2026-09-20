package ycore

import (
	"math"
	"yam/y3d"
)

var (
	FORWARD = y3d.UNIT_Z
	UP      = y3d.UNIT_Y
	RIGHT   = y3d.UNIT_X
)

type Transform struct {
	Position  y3d.Vec3
	Rotation  y3d.Quaternion
	Scale     y3d.Vec3
	Local     y3d.Mat4
	World     y3d.Mat4
	IsDirty   bool
	IsCurrent bool
}

func NewTransform() *Transform {
	return &Transform{
		Position:  y3d.Vec3{},
		Rotation:  y3d.IdenQuat(),
		Scale:     y3d.Vec3{X: 1.0, Y: 1.0, Z: 1.0},
		Local:     y3d.Identity,
		World:     y3d.Identity,
		IsCurrent: false,
	}
}

func (trans *Transform) SetScale(factor float32) {
	trans.Scale = y3d.Vec3{
		X: factor,
		Y: factor,
		Z: factor,
	}
}

func (trans *Transform) UpdateWorld(parent *Transform) {
	trans.World = parent.World.Mul(trans.Local)
}

func (trans *Transform) Recalulate() {
	trans.Local = y3d.TRS(trans.Position, trans.Rotation, trans.Scale)
}

func (trans *Transform) RecalulateNoScale() {
	m := trans.Rotation.RotMat()
	trans.Local = y3d.Mat4{
		m[0], m[1], m[2], 0,
		m[3], m[4], m[5], 0,
		m[6], m[7], m[8], 0,
		trans.Position.X, trans.Position.Y, trans.Position.Z, 1,
	}
}
func (trans *Transform) GetForward() y3d.Vec3 {
	return trans.Rotation.Rotate(FORWARD)
}

func (trans *Transform) GetRight() y3d.Vec3 {
	t := trans.Rotation.Rotate(RIGHT)
	return y3d.Normalize(t)
}

func (trans *Transform) GetUp() y3d.Vec3 {
	return trans.Rotation.Rotate(UP)
}

func (t *Transform) TransformAABB(b y3d.AABB) y3d.AABB {
	max := t.World.MulVec3(b.Max)
	min := t.World.MulVec3(b.Min)
	return y3d.AABB{
		Min: min,
		Max: max,
	}
}

func (t *Transform) TransformOBB(b y3d.OBB) y3d.OBB {
	return b.DeTransform(t.World)
}

func (t *Transform) RotateToFoward(forward y3d.Vec3) {
	dot := y3d.Dot(FORWARD, forward)
	angle := math.Acos(float64(dot))
	if dot > 0.9999 {
		t.Rotation = y3d.IdenQuat()
	} else if dot < -0.9999 {
		t.Rotation = y3d.FromAngleAxis(UP, math.Pi)
	} else {
		axis := y3d.Cross(FORWARD, forward)
		axis = y3d.Normalize(axis)
		t.Rotation = y3d.FromAngleAxis(axis, angle)
	}
	t.Recalulate()
}
