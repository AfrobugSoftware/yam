package y3d

import (
	"math"
)

type Ray struct {
	O Vec3
	D Vec3
}

func (r Ray) PointOnRay(t float32) Vec3 {
	return Add(r.O, Smul(r.D, t))
}

func (r Ray) DeTransform(m Mat4) Ray {
	r.O.X -= m[12]
	r.O.Y -= m[13]
	r.O.Z -= m[14]

	m[12] = 0
	m[13] = 0
	m[14] = 0

	(&m).Invert()
	r.O = m.MulVec3(r.O)
	r.D = m.MulVec3(r.D)

	return r
}

func (r Ray) IntersectsAABB(aabb AABB) (bool, Vec3) {
	min := aabb.Min
	max := aabb.Max
	dir := InvertVec3(r.D)
	var tmin, tmax, tYmin, tYmax, tZmin, tZmax float32
	if dir.X >= 0 {
		tmin = (min.X - r.O.X) * dir.X
		tmax = (max.X - r.O.X) * dir.X
	} else {
		tmin = (max.X - r.O.X) * dir.X
		tmax = (min.X - r.O.X) * dir.X
	}

	if dir.Y >= 0 {
		tYmin = (min.Y - r.O.Y) * dir.Y
		tYmax = (max.Y - r.O.Y) * dir.Y
	} else {
		tYmin = (max.Y - r.O.Y) * dir.Y
		tYmax = (min.Y - r.O.Y) * dir.Y
	}
	if tmin > tYmax || tYmin > tmax {
		return false, ZEROV3
	}

	if tYmin > tmin {
		tmin = tYmin
	}
	if tYmax < tmax {
		tmax = tYmax
	}

	if dir.Z >= 0 {
		tZmin = (min.Z - r.O.Z) * dir.Z
		tZmax = (max.Z - r.O.Z) * dir.Z
	} else {
		tZmin = (max.Z - r.O.Z) * dir.Z
		tZmax = (min.Z - r.O.Z) * dir.Z
	}

	if tmin > tZmax || tZmin > tmax {
		return false, ZEROV3
	}

	if tZmin > tmin {
		tmin = tZmin
	}
	if tZmax < tmax {
		tmax = tZmax
	}

	return true, Add(r.O, Smul(r.D, tmin))
}

func (r Ray) IntersectsPlane(plane Plane, cull bool) (bool, Vec3) {
	vd := Dot(plane.N, r.D)
	if math.Abs(float64(vd)) < 0.000001 {
		return false, ZEROV3
	}
	if cull && (vd > 0.0) {
		return false, ZEROV3
	}
	vo := -(Dot(plane.N, r.O) + plane.D)
	t := vo / vd
	if t < 0.0 {
		return false, ZEROV3
	}
	hit := Add(r.O, Smul(r.D, t))
	return true, hit
}

func (r Ray) IntersectsPlaneT(plane Plane, cull bool) (bool, float32) {
	vd := Dot(plane.N, r.D)
	if math.Abs(float64(vd)) < 0.000001 {
		return false, 0
	}
	if cull && (vd > 0.0) {
		return false, 0
	}
	vo := -(Dot(plane.N, r.O) + plane.D)
	t := vo / vd
	if t < 0.0 {
		return false, 0
	}
	return true, t
}

func (r Ray) IntersectsOBB(obb OBB, cull bool) (bool, float32) {
	tmin := float32(-99999.9)
	tmax := float32(+99999.9)
	cp := Sub(obb.Center, r.O)
	ex := obb.Extents.ToSlice()
	//test the 3 planes of the obb
	for i := range 3 {
		e := Dot(obb.Axes[i], cp)
		f := Dot(obb.Axes[i], r.D)

		if math.Abs(float64(f)) > 0.00001 {
			t1 := (e + ex[i]) / f
			t2 := (e - ex[i]) / f
			if t1 > t2 {
				t1, t2 = t2, t1
			}
			if t1 > tmin {
				tmin = t1
			}
			if t2 < tmax {
				tmax = t2
			}
		} else if (-e-ex[i]) > 0 || (-e+ex[i]) < 0 {
			return false, 0
		}
	}
	if tmin > 0.0 {
		return true, tmin
	}
	return true, tmax
}

func (r Ray) IntersectsTriangle(a, b, c Vec3, cull bool) (bool, float32) {
	edge1 := Sub(b, a)
	edge2 := Sub(c, a)
	pvec := Cross(r.D, edge2)
	det := Dot(edge1, pvec)
	if cull && (det < 0.0001) {
		return false, 0
	} else if det < 0.000001 && det > -0.000001 {
		return false, 0
	}
	tvec := Sub(r.O, a)
	u := Dot(tvec, pvec)
	if u < 0.0 || u > det {
		return false, 0
	}
	qvec := Cross(tvec, edge1)
	v := Dot(r.D, qvec)
	if v < 0.0 || u+v > det {
		return false, 0
	}
	t := Dot(edge2, qvec)
	fInvDet := 1.0 / det
	t *= fInvDet
	return true, t
}

func (r Ray) IntersectsSphere(s Sphere) (count float32, t [2]float32) {
	pc := Sub(r.O, s.C)
	a := Dot(r.D, r.D)
	b := 2 * Dot(r.D, pc)
	c := Dot(pc, pc) - (s.R * s.R)
	discrim := b*b - 4*a*c
	if discrim > 0 {
		t[0] = (-b + float32(math.Sqrt(float64(discrim)))) / (2 * a)
		t[1] = (-b - float32(math.Sqrt(float64(discrim)))) / (2 * a)
		count = 2
		return
	} else if discrim == 0 {
		t[0] = -b / (2 * a)
		count = 1
		return
	} else {
		count = 0
		return
	}
}
