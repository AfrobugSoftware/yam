package y3d

import (
	"math"
)

type AABB struct {
	Min Vec3
	Max Vec3
}

func (a AABB) Overlaps(b AABB) bool {
	return a.Min.X <= b.Max.X &&
		a.Max.X >= b.Min.X &&
		a.Min.Y <= b.Max.Y &&
		a.Max.Y >= b.Min.Y &&
		a.Min.Z <= b.Max.Z &&
		a.Max.Z >= b.Min.Z
}

func (b AABB) TestPlane(p Plane) bool {
	c := Smul(Add(b.Max, b.Min), 0.5)
	e := Sub(b.Max, c)

	r := math.Abs(float64(e.X*p.N.X)) +
		math.Abs(float64(e.Y*p.N.Y)) +
		math.Abs(float64(e.Z*p.N.Z))

	s := Dot(p.N, c) - p.D

	return math.Abs(float64(s)) <= r
}

func (b AABB) WhichSide(p Plane) int {
	c := Smul(Add(b.Max, b.Min), 0.5)
	e := Sub(b.Max, c)

	r := math.Abs(float64(e.X*p.N.X)) +
		math.Abs(float64(e.Y*p.N.Y)) +
		math.Abs(float64(e.Z*p.N.Z))
	projc := Dot(p.N, c)

	if projc-float32(r) > p.D {
		return 1
	} else if projc+float32(r) < p.D {
		return -1
	}
	return 0
}
