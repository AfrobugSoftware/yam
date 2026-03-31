package yecs

import (
	"yam/y3d"
)

type Transform struct {
	Position    y3d.Vec3
	Orientation y3d.Quaternion
	Scale       y3d.Vec3
	Transform   y3d.Mat4
}

func (trans *Transform) Recalulate() {
	scale := y3d.Scale(trans.Scale)
	rot := trans.Orientation.ToMat4()
	translation := y3d.Translation(trans.Position)
	trans.Transform = translation.Mul(rot).Mul(scale)
}

func (trans Transform) GetForward() y3d.Vec3 {
	return trans.Orientation.RotateVec3(y3d.UNIT_Z)
}

func (trans Transform) GetRight() y3d.Vec3 {
	return trans.Orientation.RotateVec3(y3d.UNIT_X)
}

func (trans Transform) GetUp() y3d.Vec3 {
	return trans.Orientation.RotateVec3(y3d.UNIT_Y)
}

func (t Transform) TransFormAABB(b y3d.AABB) y3d.AABB {
	world := t.Transform
	b.Max = world.MulVec3(b.Max)
	b.Min = world.MulVec3(b.Min)
	return b
}
