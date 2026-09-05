package y3d

import "math"

type OBB struct {
	Center  Vec3
	Extents Vec3 //half extents
	Axes    [3]Vec3
}

func (o OBB) project(axis Vec3) float32 {
	return float32(math.Abs(float64(Dot(o.Axes[0], axis))))*o.Extents.X +
		float32(math.Abs(float64(Dot(o.Axes[1], axis))))*o.Extents.Y +
		float32(math.Abs(float64(Dot(o.Axes[2], axis))))*o.Extents.Z
}

func (o OBB) ObbProj(axis Vec3) (max float32, min float32) {
	radius := float32(math.Abs(float64(Dot(o.Axes[0], axis))))*o.Extents.X +
		float32(math.Abs(float64(Dot(o.Axes[1], axis))))*o.Extents.Y +
		float32(math.Abs(float64(Dot(o.Axes[2], axis))))*o.Extents.Z
	d := Dot(axis, o.Center)
	return d + radius, d - radius
}

func TriProj(axis, a, b, c Vec3) (max float32, min float32) {
	min = Dot(axis, a)
	max = min
	dp := Dot(axis, b)
	if dp < min {
		min = dp
	} else if dp > max {
		max = dp
	}
	dp = Dot(axis, c)
	if dp < min {
		min = dp
	} else if dp > max {
		max = dp
	}
	return
}

func (o OBB) IntersectsTriang(a, b, c Vec3) bool {
	edge := make([]Vec3, 3)
	edge[0] = Sub(b, a)
	edge[1] = Sub(c, a)

	v := Cross(edge[0], edge[1])
	min0 := Dot(v, a)
	max0 := min0
	max1, min1 := o.ObbProj(v)
	if max1 < min0 || max0 < min1 {
		return false
	}
	//axis 1
	v = o.Axes[0]
	max0, min0 = TriProj(v, a, b, c)
	dc := Dot(v, o.Center)
	min1 = dc - o.Extents.X
	max1 = dc + o.Extents.X
	if max1 < min0 || max0 < min1 {
		return false
	}

	//axis 2
	v = o.Axes[1]
	max0, min0 = TriProj(v, a, b, c)
	dc = Dot(v, o.Center)
	min1 = dc - o.Extents.Y
	max1 = dc + o.Extents.Y
	if max1 < min0 || max0 < min1 {
		return false
	}

	//axis 3
	v = o.Axes[2]
	max0, min0 = TriProj(v, a, b, c)
	dc = Dot(v, o.Center)
	min1 = dc - o.Extents.Z
	max1 = dc + o.Extents.Z
	if max1 < min0 || max0 < min1 {
		return false
	}

	edge[2] = Sub(edge[0], edge[1])
	for i := range 3 {
		for j := range 3 {
			v = Cross(edge[i], o.Axes[j])
			max0, min0 = TriProj(v, a, b, c)
			max1, min1 = o.ObbProj(v)
			if max1 < min0 || max0 < min1 {
				return false
			}
		}
	}
	return true
}

func OBBIntersects(a, b OBB) bool {
	t := Vec3{
		X: b.Center.X - a.Center.X,
		Y: b.Center.Y - a.Center.Y,
		Z: b.Center.Z - a.Center.Z,
	}
	axes := [15]Vec3{
		a.Axes[0],
		a.Axes[1],
		a.Axes[2],
		b.Axes[0],
		b.Axes[1],
		b.Axes[2],
		Normalize(Cross(a.Axes[0], b.Axes[0])),
		Normalize(Cross(a.Axes[0], b.Axes[1])),
		Normalize(Cross(a.Axes[0], b.Axes[2])),
		Normalize(Cross(a.Axes[1], b.Axes[0])),
		Normalize(Cross(a.Axes[1], b.Axes[1])),
		Normalize(Cross(a.Axes[1], b.Axes[2])),
		Normalize(Cross(a.Axes[2], b.Axes[0])),
		Normalize(Cross(a.Axes[2], b.Axes[1])),
		Normalize(Cross(a.Axes[2], b.Axes[2])),
	}

	for _, axis := range axes {
		//can we use the sq len here instead
		if axis.Length() < NearZero {
			continue
		}
		projA := a.project(axis)
		projB := b.project(axis)
		projT := float32(math.Abs(float64(Dot(t, axis))))
		if projT > projA+projB {
			return false
		}
	}
	return true
}

func (obb OBB) DeTransform(m Mat4) OBB {
	vct := Vec3{
		X: m[12],
		Y: m[13],
		Z: m[14],
	}
	m[12], m[13], m[14] = 0.0, 0.0, 0.0
	obb.Center = m.MulVec3(obb.Center)
	obb.Axes[0] = m.MulVec3(obb.Axes[0])
	obb.Axes[1] = m.MulVec3(obb.Axes[1])
	obb.Axes[2] = m.MulVec3(obb.Axes[2])

	obb.Center = Add(obb.Center, vct)
	return obb
}

func (obb OBB) Cull(planes []Plane) int {
	for _, p := range planes {
		vN := Smul(p.N, 1.0)
		radius := obb.project(vN)
		test := Dot(vN, obb.Center) - p.D
		if test < -radius {
			return CULLED
		} else if !(test > radius) {
			return CLIPPED
		}
	}
	return VISIBLE
}
