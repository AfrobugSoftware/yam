package ygame

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

func NewTransfromation() Transform {
	return Transform{
		Position:  y3d.Vec3{},
		Rotation:  y3d.IdenQuat(),
		Scale:     y3d.Vec3{X: 1.0, Y: 1.0, Z: 1.0},
		Local:     y3d.Identity,
		World:     y3d.Identity,
		IsDirty:   false,
		IsCurrent: false,
	}
}

func (trans *Transform) Recalulate() {
	trans.Local = y3d.TRS(trans.Position, trans.Rotation, trans.Scale)
	trans.IsDirty = true
}

func (trans Transform) GetForward() y3d.Vec3 {
	return trans.Rotation.RotateVec3(FORWARD)
}

func (trans Transform) GetRight() y3d.Vec3 {
	t := trans.Rotation.RotateVec3(RIGHT)
	return y3d.Normalize(t)
}

func (trans Transform) GetUp() y3d.Vec3 {
	return trans.Rotation.RotateVec3(UP)
}

func (t Transform) TransFormAABB(b y3d.AABB) y3d.AABB {
	(&b).Scale(t.Scale)
	(&b).Translate(t.Position)
	return b
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

func (t *Transform) FromKeyFrame(k *KeyFrame) {
	t.Local = y3d.TRS(k.Position, k.Rotation, k.Scale)
	t.IsDirty = true
	t.IsCurrent = true
}
