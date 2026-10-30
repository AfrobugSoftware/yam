package ycore

import "yam/y3d"

type Box struct {
	Local y3d.AABB
	World y3d.AABB
}

func NewBox() Box {
	return Box{
		Local: y3d.UnitAABB,
		World: y3d.UnitAABB,
	}
}
func (b Box) Pick(ray y3d.Ray, fl float32) (bool, y3d.Vec3) {
	return ray.IntersectsAABB(b.World, fl)
}

func (b Box) RenderBoxMat() y3d.Mat4 {
	c := b.World.GetCenter()
	s := b.World.GetHalfSize()
	return y3d.Mat4{
		s.X, 0, 0, 0,
		0, s.Y, 0, 0,
		0, 0, s.Z, 0,
		c.X, c.Y, c.Z, 1.0,
	}
}
