package y3d

import (
	"math"
)

const (
	POINT_LIGHT = iota
	DIR_LIGHT
	SPOT_LIGHT
	AMBIENT_LIGHT
)

const (
	DEFAULT_CONSTANT  = 1.0
	DEFAULT_LINEAR    = 0.7
	DEFAULT_QUADRATIC = 1.8
	DEFAULT_DARK      = (256 / 5.0)
)

type Light struct {
	Radius float32
	Pos    Vec4
	Color  Vec4
}

func (l *Light) CalculateRadius(constant, linear, quadratic, darkIntensity float32) {
	lmax := max(l.Color.X, l.Color.Y, l.Color.Z)
	l.Radius = (-linear + float32(math.Sqrt(float64(linear*
		linear-
		4*quadratic*(constant-darkIntensity*lmax))))) /
		(2 * quadratic)
}
