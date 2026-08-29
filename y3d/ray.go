package y3d

import "math"

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

func (r Ray) IntersectsAABB(aabb AABB) (i bool, hit Vec3) {
	isInside := true
	maxT := Vec3{X: -1.0, Y: -1.0, Z: -1.0}
	//x
	if r.O.X < aabb.Min.X {
		hit.X = aabb.Min.X
		isInside = false
		if r.D.X != 0.0 {
			maxT.X = (aabb.Min.X - r.O.X) / r.D.X
		}
	} else if r.O.X > aabb.Max.X {
		hit.X = aabb.Max.X
		isInside = false
		if r.D.X != 0.0 {
			maxT.X = (aabb.Max.X - r.O.X) / r.D.X
		}
	}

	//y
	if r.O.Y < aabb.Min.Y {
		hit.Y = aabb.Min.Y
		isInside = false
		if r.D.Y != 0.0 {
			maxT.Y = (aabb.Min.Y - r.O.Y) / r.D.Y
		}
	} else if r.O.Y > aabb.Max.Y {
		hit.Y = aabb.Max.Y
		isInside = false
		if r.D.Y != 0.0 {
			maxT.Y = (aabb.Max.Y - r.O.Y) / r.D.Y
		}
	}

	//z
	if r.O.Z < aabb.Min.Z {
		hit.Z = aabb.Min.Z
		isInside = false
		if r.D.Z != 0.0 {
			maxT.Z = (aabb.Min.Z - r.O.Z) / r.D.Z
		}
	} else if r.O.Z > aabb.Max.Z {
		hit.Z = aabb.Max.Z
		isInside = false
		if r.D.Z != 0.0 {
			maxT.Z = (aabb.Max.Z - r.O.Z) / r.D.Z
		}
	}
	if isInside {
		hit = r.O
		i = true
		return
	}
	//find maximum value for maxT
	plane := 0
	maxTArray := maxT.ToSlice()
	if maxT.Y > maxTArray[plane] {
		plane = 1
	}
	if maxT.Z > maxTArray[plane] {
		plane = 2
	}
	if maxTArray[plane] < 0.0 {
		i = false
		return
	}
	if plane != 0 {
		hit.X = r.O.X + maxT.X*r.D.X
		if hit.X < aabb.Min.X-0.00001 || hit.X < aabb.Max.X+0.00001 {
			i = false
			return
		}
	}
	if plane != 1 {
		hit.Y = r.O.Y + maxT.Y*r.D.Y
		if hit.Y < aabb.Min.Y-0.00001 || hit.Y < aabb.Max.Y+0.00001 {
			i = false
			return
		}
	}
	if plane != 2 {
		hit.Z = r.O.Z + maxT.Z*r.D.Z
		if hit.Z < aabb.Min.Z-0.00001 || hit.Z < aabb.Max.Z+0.00001 {
			i = false
			return
		}
	}
	i = true
	return
}

func (r Ray) IntersectsPlane(plane Plane, cull bool) (bool, Vec3) {
	vd := Dot(plane.N, r.D)
	if math.Abs(float64(vd)) < 0.000001 {
		return false, ZEROV
	}
	if cull && (vd > 0.0) {
		return false, ZEROV
	}
	vo := -(Dot(plane.N, r.O) + plane.D)
	t := vo / vd
	if t < 0.0 {
		return false, ZEROV
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
