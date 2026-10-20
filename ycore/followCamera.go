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
	OwnersPosition y3d.Vec3
	OwnersForward  y3d.Vec3
	targetPoint    y3d.Vec3
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
	target := y3d.Add(ff.OwnersPosition, y3d.Smul(ff.OwnersForward, ff.TargetDistance))

	dir := y3d.Normalize(y3d.Sub(target, ff.Position))
	vRight := y3d.Normalize(y3d.Cross(y3d.UNIT_Y, dir))
	vUp := y3d.Cross(dir, vRight)

	ff.Up = vUp
	ff.Right = vRight
	ff.Dir = dir
}

func (ff *FollowCamera) Update(dt float32) {
	ff.RecalAxis() // depends on the target we are following
}
