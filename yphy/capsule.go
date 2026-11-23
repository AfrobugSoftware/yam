package yphy

import (
	"bytes"
	"fmt"
	"yam/y3d"
)

type Capsule struct {
	Local y3d.Capsule
	World y3d.Capsule
}

func BoundCapsule(stride, relativeOffset int, dataV, dataI *bytes.Buffer) Capsule {
	return Capsule{}
}

func (c Capsule) Transform(m y3d.Mat4) Collider {
	c.World = c.Local.Transform(m)
	return c
}

func (cp Capsule) Intersects(c Collider) bool {
	return false
}

func (cp Capsule) GenerateContact(c Collider) Contact {
	switch t := c.(type) {
	case Box:
		fmt.Println(t.World)
	}

	return Contact{}
}

func (cp Capsule) Pick(ray y3d.Ray) (bool, y3d.Vec3) {

	return false, y3d.Vec3{}
}

func (cp Capsule) GetColliderType() ColliderType {
	return COLLIDER_CAPSULE
}
