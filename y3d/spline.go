package y3d

import "math"

type Spline struct {
	ControlPoints []Vec3
}

// equation: p(t) = 0.5 * (2p1 + (-p0 + p2)t + (2p0 - 5p1 + 4p2 - p3)t^3)
func (s Spline) Compute(startIdx int, t float32) Vec3 {
	l := len(s.ControlPoints)
	if startIdx == 0 {
		return s.ControlPoints[0]
	} else if startIdx >= l {
		return s.ControlPoints[l-1]
	} else if startIdx+2 >= l {
		return s.ControlPoints[l-1]
	}
	p0 := s.ControlPoints[startIdx-1]
	p1 := s.ControlPoints[startIdx]
	p2 := s.ControlPoints[startIdx+1]
	p3 := s.ControlPoints[startIdx+2]
	t3 := math.Pow(float64(t), 3)

	a := Smul(p1, 2)
	b := Smul(Add(NegateVec3(p0), p2), float32(t))

	c := Sub(Smul(p0, 2), Smul(p1, -5))
	d := Sub(Smul(p2, 4), p3)
	e := Smul(Add(c, d), float32(t3))

	f := Smul(Add(Add(a, b), e), 0.5)
	return f
}
