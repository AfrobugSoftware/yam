package y3d

import "math"

type Sphere struct {
	C Vec3
	R float32
}

func SphereIntersects(a, b Sphere) bool {
	d := DistanceSqured(b.C, a.C)
	return d <= (a.R+b.R)*(a.R+b.R)
}

func (s Sphere) Contains(point Vec3) bool {
	distsq := DistanceSqured(s.C, point)
	return distsq <= (s.R * s.R)
}
func (s Sphere) WhichSide(p Plane) Side {
	d := p.SignedDistance(s.C)
	switch {
	case d > s.R:
		return FRONT
	case d < -s.R:
		return BACK
	}
	return INTERSECTS
}

//finds, if possible the exact point in which two spheres intersects between 2 frames
func SweptSphere(p0, p1, q0, q1 Sphere) (float32, bool) {
	X := Sub(p0.C, q0.C)
	Y := Sub(Sub(p1.C, p0.C), Sub(q1.C, q0.C))
	a := Dot(Y, Y)
	b := 2 * Dot(X, Y)
	sumRadaii := p0.R + q0.R
	c := Dot(X, X) - (sumRadaii * sumRadaii)
	disc := b*b - 4.0*a*c
	if disc < 0 {
		return 0, false
	}
	disc = float32(math.Sqrt(float64(disc)))
	outT := -(b - disc) / 2 * a
	return outT, (outT >= 0 && outT < 1)
}
