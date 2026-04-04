package y3d

import (
	"math"
	"unsafe"
)

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

func FastInvSqrt(value float32) float32 {
	half := 0.5 * value
	i := *(*int)(unsafe.Pointer(&value))
	i = 0x5f3759df - (i >> 1)
	value = *(*float32)(unsafe.Pointer(&i))
	value = value * (1.5 - half*value*value)
	return value
}
