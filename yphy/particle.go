package yphy

import "yam/y3d"

type Particle struct {
	Position    y3d.Vec3
	Velocity    y3d.Vec3
	InverseMass float32
}
