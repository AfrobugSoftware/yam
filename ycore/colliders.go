package ycore

import "yam/y3d"

//collider types

const (
	COLLIDER_SPHERE = iota
)

type Collider interface {
	Pick(r y3d.Ray) (bool, y3d.Vec3)
}
