package yecs

import "yam/y3d"

type Transform struct {
	Position    y3d.Vec3
	Orientation y3d.Quaternion
	Scale       y3d.Vec3

	NeedCalculation bool
	cacheTransform  y3d.Mat4
}

func (trans Transform) GetTransformation() y3d.Mat4 {
	if trans.NeedCalculation {
		trans.cacheTransform = y3d.Scale(trans.Scale).
			Mul(trans.Orientation.ToMat4()).
			Mul(y3d.Translation(trans.Position))
	}
	return trans.cacheTransform
}
