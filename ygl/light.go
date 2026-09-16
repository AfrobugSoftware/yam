package ygl

import (
	"yam/y3d"
)

const (
	POINT_LIGHT = iota
	DIR_LIGHT
	SPOT_LIGHT
	AMBIENT_LIGHT
)

type Light struct {
	Pos    y3d.Vec4
	Color  y3d.Vec4
	Volume y3d.Mat4
}

func (l *Light) CalcOmniLightMat(radius float32) {
	invRadius := 0.5 / radius
	ms := y3d.Scale(y3d.Vec3{
		X: invRadius,
		Y: invRadius,
		Z: invRadius,
	})
	mT := y3d.Translation(
		y3d.Vec3{
			X: -l.Pos.X,
			Y: -l.Pos.Y,
			Z: -l.Pos.Z,
		},
	)
	mB2 := y3d.Translation(
		y3d.Vec3{
			X: 0.5,
			Y: 0.5,
			Z: 0.5,
		},
	)
	l.Volume = mT.Mul(ms).Mul(mB2)
}
