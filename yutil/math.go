package yutil

import "math"

const (
	DEGTORAD = 180.0 / math.Pi
	RADTODEG = math.Pi / 180.0
)

func ToDegree(rad float64) float64 {
	return rad * DEGTORAD
}

func ToRadians(deg float64) float64 {
	return deg * RADTODEG
}
