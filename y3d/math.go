package y3d

import "math"

const (
	DEGTORAD = 180.0 / math.Pi
	RADTODEG = math.Pi / 180.0
	NearZero = 1e-10
)

func ToDegree(rad float64) float64 {
	return rad * DEGTORAD
}

func ToRadians(deg float64) float64 {
	return deg * RADTODEG
}

func Clamp(a, max, min float32) float32 {
	if a > max {
		a = max
	} else if a < min {
		a = min
	}
	return a
}
