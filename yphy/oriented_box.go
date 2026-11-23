package yphy

import (
	"bytes"
	"fmt"
	"yam/y3d"
)

type OrientedBox struct {
	Local y3d.OBB
	World y3d.OBB
}

func BoundOrientedBox(stride, relativeOffset int, dataV, dataI *bytes.Buffer) OrientedBox {
	return OrientedBox{}
}

func (b OrientedBox) Intersects(c Collider) bool {
	return false
}

func (b OrientedBox) Transform(m y3d.Mat4) Collider {
	b.World = b.Local.Transform(m)
	return b
}

func (b OrientedBox) GenerateContact(c Collider) Contact {
	switch t := c.(type) {
	case Box:
		fmt.Println(t.World)
	}

	return Contact{}
}

func (b OrientedBox) Pick(ray y3d.Ray) (bool, y3d.Vec3) {
	bt, t := ray.IntersectsOBB(b.World, false)
	if !bt {
		return false, y3d.Vec3{}
	}
	point := ray.O.AddScale(ray.D, t)
	return bt, point
}

func (b OrientedBox) GetColliderType() ColliderType {
	return COLLIDER_OBB
}
