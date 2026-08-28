package y3d

type Ray struct {
	O Vec3
	D Vec3
}

func (r Ray) PointOnRay(t float32) Vec3 {
	return Add(r.O, Smul(r.D, t))
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
