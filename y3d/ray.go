package y3d

type Ray struct {
	O Vec3
	D Vec3
}

func (r Ray) PointOnRay(t float32) Vec3 {
	return Add(r.O, Smul(r.D, t))
}
