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
	Local y3d.Mat4
	World y3d.Mat4
}

func NewTransform() *Transform {
	return &Transform{
		Local: y3d.Identity,
		World: y3d.Identity,
	}
}

func (trans *Transform) UpdateWorld(parent *Transform) {
	trans.World = parent.World.Mul(trans.Local)
}

func (t *Transform) TransformAABB(b y3d.AABB) y3d.AABB {
	translate := y3d.Vec3{
		X: t.World[12],
		Y: t.World[13],
		Z: t.World[14],
	}
	max := y3d.Add(b.Max, translate)
	min := y3d.Add(b.Min, translate)
	return y3d.AABB{
		Min: min,
		Max: max,
	}
}

func (t *Transform) TransformOBB(b y3d.OBB) y3d.OBB {
	return b.DeTransform(t.World)
}

func RotateToFoward(forward y3d.Vec3) y3d.Quaternion {
	dot := y3d.Dot(FORWARD, forward)
	angle := math.Acos(float64(dot))
	var rot y3d.Quaternion
	if dot > 0.9999 {
		rot = y3d.IdenQuat()
	} else if dot < -0.9999 {
		rot = y3d.FromAngleAxis(UP, math.Pi)
	} else {
		axis := y3d.Cross(FORWARD, forward)
		axis = y3d.Normalize(axis)
		rot = y3d.FromAngleAxis(axis, float32(angle))
	}
	return rot
}
