package ycontroller

import "yam/y3d"

type Controller interface {
	Update(deltatime float32)
	GetDir() y3d.Vec3
	GetUp() y3d.Vec3
	GetRight() y3d.Vec3
	GetPosition() y3d.Vec3
	GetOrientation() y3d.Quaternion
}
