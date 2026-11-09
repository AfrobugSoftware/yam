package yphy

import (
	"yam/y3d"
)

// collider types
type ColliderType int

const (
	COLLIDER_SPHERE ColliderType = iota
	COLLIDER_AABB
	COLLIDER_CAPSULE
	COLLIDER_OBB
	COLLIDER_PLANE
)

type Collider interface {
	Pick(r y3d.Ray) (bool, y3d.Vec3)
	Intersects(c Collider) bool
	Transform(w y3d.Mat4) Collider
	GetColliderType() ColliderType
}
