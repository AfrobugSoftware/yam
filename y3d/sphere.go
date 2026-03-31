package y3d

import "math"

type Sphere struct {
	C Vec3
	R float32
}

func SphereIntersects(a, b Sphere) bool {
	d := float64(DistanceSqured(b.C, a.C))
	if d <= math.Pow(float64(a.R+b.R), 2) {
		return true
	} else {
		return false
	}
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
