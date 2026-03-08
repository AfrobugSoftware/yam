package y3d

import "math"

type Sphere struct {
	C Vec3
	R float64
}

func SphereIntersects(a, b Sphere) bool {
	d := DistanceSqured(b.C, a.C)
	if d <= math.Pow(a.R+b.R, 2) {
		return true
	} else {
		return false
	}
}
