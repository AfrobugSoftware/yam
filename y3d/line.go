package y3d

import (
	"math"
	"slices"
)

type LineSegment struct {
	Start Vec3
	End   Vec3
}

func (ls LineSegment) PointOnLine(t float32) Vec3 {
	return Add(ls.Start, Smul(Sub(ls.End, ls.Start), t))
}

func (ls LineSegment) MinDistSq(point Vec3) float32 {
	ab := Sub(ls.End, ls.Start)
	ba := NegateVec3(ab)
	ac := Sub(point, ls.Start)
	bc := Sub(point, ls.End)

	if Dot(ab, ac) < 0.0 {
		return ac.LengthSq()
	} else if Dot(ba, bc) < 0.0 {
		return bc.LengthSq()
	} else {
		//compute p
		scalerProj := Dot(ac, ab) / Dot(ab, ab)
		p := Smul(ab, scalerProj)
		return DistanceSqured(ac, p)
	}
}

func MinLineSegmentDistSq(p, q LineSegment) float32 {
	d1 := Sub(p.End, p.Start)
	d2 := Sub(p.End, q.Start)
	r := Sub(p.Start, q.Start)

	a := Dot(d1, d1)
	e := Dot(d2, d2)
	f := Dot(d2, r)

	var s, t float32
	switch {
	case a <= NearZero && e <= NearZero:
		return r.LengthSq()
	case a <= NearZero:
		s = 0
		t = Clamp(f/e, 0, 1)
	default:
		c := Dot(d1, r)
		if e <= NearZero {
			t = 0
			s = Clamp(-c/a, 0, 1)
		} else {
			b := Dot(d1, d2)
			denom := a*e - b*b
			if denom > NearZero {
				s = Clamp((b*f-c*e)/denom, 0, 1)
			} else {
				s = 0
			}
			t = (b*s + f) / e
			if t < 0 {
				t = 0
				s = Clamp(-c/a, 0, 1)
			} else if t > 1 {
				t = 1
				s = Clamp((b-c)/a, 0, 1)
			}
		}
	}
	closestOnP := Add(p.Start, Smul(d1, s))
	closestOnQ := Add(q.Start, Smul(d2, t))

	return (Sub(closestOnP, closestOnQ)).LengthSq()
}

func (l LineSegment) IntersectsPlane(p Plane) (float32, bool) {
	d1 := Sub(l.End, l.Start)
	denom := Dot(d1, p.N)
	if denom < NearZero {
		if Dot(l.Start, p.N)-p.D < NearZero {
			return 0, true
		} else {
			return 0, false
		}
	} else {
		number := -Dot(l.Start, p.N) - p.D
		outT := number / denom
		intersect := outT >= 0 && outT < 1
		return outT, intersect
	}
}

func (l LineSegment) IntersectSphere(s Sphere) (float32, bool) {
	X := Sub(l.Start, s.C)
	Y := Sub(l.End, l.Start)

	a := Dot(Y, Y)
	b := 2 * Dot(X, Y)
	c := Dot(X, X) - (s.R * s.R)

	disc := (b * b) - 4*a*c
	if disc < 0 {
		return 0, false
	}
	disc = float32(math.Sqrt(float64(disc)))
	tMin := (-b - disc) / 2.0 * a
	tMax := (-b + disc) / 2.0 * a
	if tMin > 0 && tMin < 1 {
		return tMin, true
	} else if tMax > 0 && tMax < 1 {
		return tMax, true
	}
	return 0, false
}

func TestSidePlane(start, end, negD float32, out []float32) bool {
	denom := end - start
	if denom < NearZero {
		return false
	}
	num := -start + negD
	t := num / denom
	if t >= 0 && t <= 1 {
		out = append(out, t)
		return true
	}
	return false
}

func (l LineSegment) IntersectAABB(b AABB) (float32, bool) {
	ts := make([]float32, 6)
	TestSidePlane(l.Start.X, l.End.X, b.Min.X, ts)
	TestSidePlane(l.Start.X, l.End.X, b.Max.X, ts)

	TestSidePlane(l.Start.Y, l.End.Y, b.Min.Y, ts)
	TestSidePlane(l.Start.Y, l.End.Y, b.Max.Y, ts)

	TestSidePlane(l.Start.Z, l.End.Z, b.Min.Z, ts)
	TestSidePlane(l.Start.Z, l.End.Z, b.Max.Z, ts)

	slices.Sort(ts)

	var point Vec3
	for _, t := range ts {
		point = l.PointOnLine(t)
		if b.Contains(point) {
			return t, true
		}
	}
	return 0, false
}

func (l LineSegment) GetHitNormal(box AABB, t float32) Vec3 {
	dir := Sub(l.End, l.Start)
	hit := Add(l.Start, Smul(dir, t))
	halfSize := box.GetHalfSize()
	local := Sub(hit, box.GetCenter())
	nx := float64(local.X / halfSize.X)
	ny := float64(local.Y / halfSize.Y)
	nz := float64(local.Z / halfSize.Z)

	ax, ay, az := math.Abs(nx), math.Abs(ny), math.Abs(nz)

	switch {
	case ax >= ay && ax >= az:
		return Vec3{float32(nx), 0, 0}
	case ay >= ax && ay >= az:
		return Vec3{0, float32(ny), 0}
	default:
		return Vec3{0, 0, float32(nz)}
	}

}
