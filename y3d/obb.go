package y3d

import "math"

type OBB struct {
	Center  Vec3
	Extents Vec3
	Axes    [3]Vec3
}

// project projects the OBB onto an axis and returns the radius of the projection
func (o *OBB) project(axis Vec3) float64 {
	return math.Abs(Dot(o.Axes[0], axis))*o.Extents.X +
		math.Abs(Dot(o.Axes[1], axis))*o.Extents.Y +
		math.Abs(Dot(o.Axes[2], axis))*o.Extents.Z
}

// OBBIntersects returns true if the two OBBs are intersecting
// using the Separating Axis Theorem (SAT) with 15 axes
func OBBIntersects(a, b OBB) bool {
	t := Vec3{
		X: b.Center.X - a.Center.X,
		Y: b.Center.Y - a.Center.Y,
		Z: b.Center.Z - a.Center.Z,
	}

	// 15 potential separating axes:
	// - 3 axes of OBB A
	// - 3 axes of OBB B
	// - 9 cross products of each axis pair
	axes := [15]Vec3{
		// Face axes of A
		a.Axes[0],
		a.Axes[1],
		a.Axes[2],
		// Face axes of B
		b.Axes[0],
		b.Axes[1],
		b.Axes[2],
		// Edge cross products
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
		// Skip near-zero axes (degenerate cross products from parallel edges)
		if axis.Length() < 1e-10 {
			continue
		}

		// Project both OBBs and the translation onto this axis
		projA := a.project(axis)
		projB := b.project(axis)
		projT := math.Abs(Dot(t, axis))

		// If there is a gap, we found a separating axis — no intersection
		if projT > projA+projB {
			return false
		}
	}

	// No separating axis found — OBBs are intersecting
	return true
}
