package ycore

import (
	"yam/y3d"
	"yam/ycontroller"
)

type FollowCamera struct {
	ycontroller.MovementController
	HDistance      float32
	VDistance      float32
	TargetDistance float32
	TargetPoint    y3d.Vec3
	TargetDir      y3d.Vec3
	OwnersPosition y3d.Vec3
}

func NewFollowCamera(pos y3d.Vec3, hdist, vdist, targetdist float32) *FollowCamera {
	return &FollowCamera{
		HDistance:          hdist,
		VDistance:          vdist,
		TargetDistance:     targetdist,
		MovementController: *ycontroller.NewMovementController(pos, y3d.IdenQuat()),
	}
}

func (ff *FollowCamera) RecalAxis() {
	v := y3d.Smul(ff.Dir, -ff.HDistance)
	u := y3d.Smul(y3d.UNIT_Y, ff.VDistance)
	ff.Position = y3d.Add(ff.Position, v, u)
	ff.TargetPoint = y3d.Add(ff.OwnersPosition, y3d.Smul(ff.TargetDir, ff.TargetDistance))

	dir := y3d.Normalize(y3d.Sub(ff.TargetPoint, ff.Position))
	up := y3d.UNIT_Y

	dot := y3d.Dot(dir, up)
	tmp := y3d.Smul(up, dot)
	vUp := y3d.Sub(up, tmp)
	l := vUp.Length()
	if l < y3d.NearZero {
		vY := y3d.Vec3{X: 0, Y: 1, Z: 0}
		tmp = y3d.Smul(dir, dir.Y)
		vUp = y3d.Sub(vY, tmp)
		l = vUp.Length()
		if l < y3d.NearZero {
			vY := y3d.Vec3{X: 0, Y: 0, Z: 1}
			tmp = y3d.Smul(dir, dir.Z)
			vUp = y3d.Sub(vY, tmp)
			l = vUp.Length()
			if l < y3d.NearZero {
				panic("LookAt: up vector is parallel to direction vector")
			}
		}
	}
	vUp = y3d.Smul(vUp, 1.0/l)
	vRight := y3d.Normalize(y3d.Cross(vUp, dir))

	ff.Up = vUp
	ff.Right = vRight
	ff.Dir = dir
}

func (ff *FollowCamera) Update(dt float32) {
	ff.RecalAxes() // depends on the target we are following
}
