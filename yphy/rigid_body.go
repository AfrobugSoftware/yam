package yphy

import "yam/y3d"

type RigidBody struct {
	InverseMass   float32
	LinearDamping float32
	Position      y3d.Vec3
	Orientation   y3d.Vec3
	Velocity      y3d.Vec3
	Rotation      y3d.Vec3

	Transfrom y3d.Mat4
}
