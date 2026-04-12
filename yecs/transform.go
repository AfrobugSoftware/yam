package yecs

import (
	"math"
	"yam/y3d"
)

type Transform struct {
	Position  y3d.Vec3
	Rotation  y3d.Quaternion
	Scale     y3d.Vec3
	Transform y3d.Mat4
}

func (trans *Transform) Recalulate() {
	scale := y3d.Scale(trans.Scale)
	rot := trans.Rotation.ToMat4()
	translation := y3d.Translation(trans.Position)
	trans.Transform = translation.Mul(rot).Mul(scale)
}

func (trans Transform) GetForward() y3d.Vec3 {
	return trans.Rotation.RotateVec3(y3d.UNIT_Z)
}

func (trans Transform) GetRight() y3d.Vec3 {
	return trans.Rotation.RotateVec3(y3d.UNIT_X)
}

func (trans Transform) GetUp() y3d.Vec3 {
	return trans.Rotation.RotateVec3(y3d.UNIT_Y)
}

func (t Transform) TransFormAABB(b y3d.AABB) y3d.AABB {
	(&b).Scale(t.Scale)
	//no rotation yet.
	(&b).Translate(t.Position)
	return b
}

func (t *Transform) RotateToFoward(forward y3d.Vec3) {
	dot := y3d.Dot(y3d.UNIT_Z, forward)
	angle := math.Acos(float64(dot))
	if dot > 0.9999 {
		t.Rotation = y3d.IdenQuat()
	} else if dot < -0.9999 {
		t.Rotation = y3d.FromAngleAxis(y3d.UNIT_Y, math.Pi)
	} else {
		axis := y3d.Cross(y3d.UNIT_Z, forward)
		axis = y3d.Normalize(axis)
		t.Rotation = y3d.FromAngleAxis(axis, angle)
	}
	t.Recalulate()
}
