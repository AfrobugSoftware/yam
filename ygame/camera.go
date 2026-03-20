package ygame

import "yam/y3d"

type Camara struct {
	Pos    y3d.Vec3
	Vel    y3d.Vec3
	Follow y3d.Vec3
	LookAt y3d.Vec3
	View   y3d.Mat4
}
