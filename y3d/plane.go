package y3d

import "math"

type Side int

const (
	BACK    Side = -1
	PLANER  Side = 0
	FRONT   Side = 1
	CLIPPED      = 2
	CULLED       = 3
	VISIBLE      = 4
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
func PlaneFromPointNormal(p, normal Vec3) Plane {
	return Plane{
		N: normal,
		D: -(Dot(normal, p)),
	}
}

func (p Plane) SignedDistance(point Vec3) float32 {
	return Dot(p.N, point) - p.D
}

func (p Plane) Classify(v Vec3) Side {
	f := Dot(p.N, v) - p.D
	switch {
	case f > NearZero:
		return FRONT
	case f < NearZero:
		return BACK
	default:
		return PLANER
	}
}

func (p Plane) ClassifyPolygon(polygon Polygon) Side {
	s := p.Classify(polygon.Points[0])
	for i := 1; i < len(polygon.Points); i++ {
		s1 := p.Classify(polygon.Points[i])
		if s != s1 {
			return CLIPPED
		}
	}
	return s
}

func (p Plane) IntersectTriangle(a, b, c Vec3) bool {
	n := p.Classify(a)
	if n == p.Classify(b) && n == p.Classify(c) {
		return false
	}
	return true
}

func (p Plane) IntersectsPlane(p2 Plane, r *Ray) bool {
	cross := Cross(p.N, p2.N)
	sqLen := cross.LengthSq()
	if sqLen < 1e-08 {
		return false
	}
	if r != nil {
		n00 := p.N.LengthSq()
		n01 := Dot(p.N, p2.N)
		n11 := p2.N.LengthSq()
		det := n00*n11 - n01*n01
		if math.Abs(float64(det)) < 1e-08 {
			return false
		}
		invDet := 1.0 / det
		c0 := (n11*p.D - n01*p2.D) * invDet
		c1 := (n00*p2.D - n01*p.D) * invDet
		r.D = cross
		r.O = Add(Smul(p.N, c0), Smul(p2.N, c1))
	}
	return true
}

func (p Plane) IntersectsAABB(aabb AABB) bool {
	var min, max Vec3
	if p.N.X >= 0 {
		min.X = aabb.Min.X
		max.X = aabb.Max.X
	} else {
		min.X = aabb.Max.X
		max.X = aabb.Min.X
	}
	if p.N.Y >= 0 {
		min.Y = aabb.Min.Y
		max.Y = aabb.Max.Y
	} else {
		min.Y = aabb.Max.Y
		max.Y = aabb.Min.Y
	}
	if p.N.Z >= 0 {
		min.Z = aabb.Min.Z
		max.Z = aabb.Max.Z
	} else {
		min.Z = aabb.Max.Z
		max.Z = aabb.Min.Z
	}
	if n := Dot(p.N, min) + p.D; n > 0 {
		return false
	}
	if n := Dot(p.N, max) + p.D; n >= 0 {
		return true
	}
	return false
}

func (p Plane) IntersectsOBB(obb OBB) bool {
	radius := obb.project(p.N)
	distance := math.Abs(float64(p.SignedDistance(obb.Center)))
	return distance <= float64(radius)
}
