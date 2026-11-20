package ygame

import (
	"math"
	"yam/y3d"
)

type FreeRoamCamera struct {
	Position y3d.Vec3
	Up       y3d.Vec3
	Front    y3d.Vec3
	Right    y3d.Vec3
	Pitch    float32
	Yaw      float32
	Row      float32
}

func NewCamera() *FreeRoamCamera {
	return &FreeRoamCamera{}
}

func (frc *FreeRoamCamera) UpdateMouseMovement(m y3d.Vec2) {
	if math.Abs(float64(m.X)) > y3d.NearZero ||
		math.Abs(float64(m.Y)) > y3d.NearZero {
		frc.Yaw += float32(y3d.ToRadians(float64(m.X))) * 0.5
		frc.Pitch += float32(y3d.ToRadians(float64(m.Y))) * 0.5
		if frc.Pitch > float32(y3d.ToRadians(89.0)) {
			frc.Pitch = float32(y3d.ToRadians(89.0))
		} else if frc.Pitch < -float32(y3d.ToRadians(89.0)) {
			frc.Pitch = -float32(y3d.ToRadians(89.0))
		}
		frc.Front = y3d.Normalize(y3d.Vec3{
			X: float32(math.Cos(float64(frc.Yaw)) * math.Cos(float64(frc.Pitch))),
			Y: float32(math.Sin(float64(frc.Pitch))),
			Z: float32(math.Sin(float64(frc.Yaw)) * math.Cos(float64(frc.Pitch))),
		})
	}
}

func (frc *FreeRoamCamera) Update() y3d.Mat4 {
	dir := frc.Front
	vRight := y3d.Normalize(y3d.Cross(y3d.UNIT_Y, dir))
	vUp := y3d.Cross(dir, vRight)
	eye := frc.Position
	frc.Up = vUp
	frc.Front = dir
	frc.Right = vRight
	return y3d.Mat4{
		vRight.X, vUp.X, dir.X, 0,
		vRight.Y, vUp.Y, dir.Y, 0,
		vRight.Z, vUp.Z, dir.Z, 0,
		-y3d.Dot(vRight, eye), -y3d.Dot(vUp, eye), -y3d.Dot(dir, eye), 1,
	}
}
