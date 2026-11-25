package yphy

import "yam/y3d"

type Plane struct {
	Local y3d.Plane
	World y3d.Plane
}

func (p Plane) Pick(r y3d.Ray) (bool, y3d.Vec3) {
	return r.IntersectsPlane(p.World)
}
func (p Plane) Intersects(c Collider) bool {
	return false
}
func (p Plane) Transform(w y3d.Mat4) Collider {
	i := w
	i.Invert()
	i.Transpose()
	n := i.MulVec3(p.Local.N)
	o := y3d.Smul(p.Local.N, p.Local.D)
	o4 := w.MulVec4(y3d.Vec4{
		X: o.X,
		Y: o.Y,
		Z: o.Z,
		W: 1,
	})
	p.World.N = n
	p.World.D = y3d.Dot(y3d.Vec3{
		X: o4.X,
		Y: o4.Y,
		Z: o4.Z,
	}, n)
	return p
}
func (p Plane) GetColliderType() ColliderType {
	return COLLIDER_PLANE
}
func (p Plane) GenerateContact(c Collider, cq *ContactQueue) {
}
