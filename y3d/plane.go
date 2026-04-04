package y3d

type Side int

const (
	BACK       Side = -1
	INTERSECTS Side = 0
	FRONT      Side = 1
)

type Plane struct {
	N Vec3
	D float32
}

func NewPlane(a, b, c Vec3) Plane {
	ab := Sub(b, a)
	ac := Sub(c, a)
	norm := Normalize(Cross(ab, ac))
	return Plane{
		N: norm,
		D: -(Dot(norm, a)),
	}
}
func (p Plane) SignedDistance(point Vec3) float32 {
	return Dot(p.N, point) - p.D
}
