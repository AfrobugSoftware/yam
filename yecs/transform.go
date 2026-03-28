package yecs

import (
	"yam/y3d"
)

type Transform struct {
	Position        y3d.Vec3
	Orientation     y3d.Quaternion
	Scale           y3d.Vec3
	Name            string
	NeedCalculation bool
	cacheTransform  y3d.Mat4
}

func (trans Transform) GetTransformation() y3d.Mat4 {
	if trans.NeedCalculation {
		scale := y3d.Scale(trans.Scale)
		rot := trans.Orientation.ToMat4()
		translation := y3d.Translation(trans.Position)
		trans.cacheTransform = translation.Mul(rot).Mul(scale)
		trans.NeedCalculation = false
	}
	return trans.cacheTransform
}

func (trans Transform) GetForward() y3d.Vec3 {
	return trans.Orientation.RotateVec3(y3d.UNIT_Z)
}
