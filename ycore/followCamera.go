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
	EntityPosition y3d.Vec3
	EntityForward  y3d.Vec3
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

// have to look into this, this does not seem to work
// or rather is is confusing
// the start position of the camara seem to still be at the origin
// if we use the offsets on the unit x and y axis
// this supposed to be relative to the position of the entity target
func (ff *FollowCamera) RecalAxis() {
	v := y3d.Smul(y3d.UNIT_X, -ff.HDistance)
	u := y3d.Smul(y3d.UNIT_Y, ff.VDistance)
	ff.Position = y3d.Add(ff.Position, v, u)
	target := y3d.Add(ff.EntityPosition, y3d.Smul(ff.EntityForward, ff.TargetDistance))

	dir := y3d.Normalize(y3d.Sub(target, ff.Position))
	vRight := y3d.Normalize(y3d.Cross(y3d.UNIT_Y, dir))
	vUp := y3d.Cross(dir, vRight)

	ff.Up = vUp
	ff.Right = vRight
	ff.Dir = dir
}

func (ff *FollowCamera) UpdateEntity(pos, dir y3d.Vec3) {
	ff.EntityPosition = pos
	ff.EntityForward = dir
}

func (ff *FollowCamera) Update(dt float32) {
	ff.RecalAxis() // depends on the target we are following
}
