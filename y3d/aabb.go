package y3d

import (
	"math"
)

type AABB struct {
	Min Vec3
	Max Vec3
}

func (a AABB) Overlaps(b AABB) bool {
	no := (a.Max.X < b.Min.X ||
		a.Max.Y < b.Min.Y ||
		a.Max.Z < b.Min.Z ||
		b.Max.X < a.Min.X ||
		b.Max.Y < a.Min.Y ||
		b.Max.Z < a.Min.Z)
	return !no
}

func (box AABB) GetCenter() Vec3 {
	return Vec3{
		(box.Min.X + box.Max.X) * 0.5,
		(box.Min.Y + box.Max.Y) * 0.5,
		(box.Min.Z + box.Max.Z) * 0.5,
	}
}

func (box AABB) GetHalfSize() Vec3 {
	return Vec3{
		(box.Max.X - box.Min.X) * 0.5,
		(box.Max.Y - box.Min.Y) * 0.5,
		(box.Max.Z - box.Min.Z) * 0.5,
	}
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
	c := b.GetCenter()
	e := b.GetHalfSize()

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

func (b AABB) Contains(point Vec3) bool {
	outside := (point.X < b.Min.X ||
		point.Y < b.Min.Y ||
		point.Z < b.Min.Z ||
		point.X > b.Max.X ||
		point.Y > b.Max.Y ||
		point.Z > b.Max.Z)
	return !outside
}

func (b *AABB) Translate(pos Vec3) {
	b.Max = Add(b.Max, pos)
	b.Min = Add(b.Min, pos)
}

func (b *AABB) Scale(scale Vec3) {
	b.Max = Mul(b.Max, scale)
	b.Min = Mul(b.Min, scale)
}

func (b AABB) MinDistSq(point Vec3) float32 {
	dx := max((b.Min.X - point.X), 0.0)
	dx = max(dx, (point.X - b.Max.X))

	dy := max((b.Min.Y - point.Y), 0.0)
	dy = max(dy, (point.Y - b.Max.Y))

	dz := max((b.Min.Z - point.Z), 0.0)
	dz = max(dz, (point.Z - b.Max.Z))

	return (dx * dx) + (dy * dy) + (dz * dz)
}

func (b AABB) IntersectsSphere(sphere Sphere) bool {
	distSq := b.MinDistSq(sphere.C)
	return distSq <= (sphere.R * sphere.R)
}
