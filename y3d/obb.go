package y3d

import "math"

type OBB struct {
	Center  Vec3
	Extents Vec3
	Axes    [3]Vec3
}

func (o *OBB) project(axis Vec3) float32 {
	return float32(math.Abs(float64(Dot(o.Axes[0], axis))))*o.Extents.X +
		float32(math.Abs(float64(Dot(o.Axes[1], axis))))*o.Extents.Y +
		float32(math.Abs(float64(Dot(o.Axes[2], axis))))*o.Extents.Z
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
		if axis.Length() < NearZero {
			continue
		}

		// Project both OBBs and the translation onto this axis
		projA := a.project(axis)
		projB := b.project(axis)
		projT := float32(math.Abs(float64(Dot(t, axis))))

		// If there is a gap, we found a separating axis — no intersection
		if projT > projA+projB {
			return false
		}
	}

	// No separating axis found — OBBs are intersecting
	return true
}
