package yecs

import "yam/y3d"

const (
	POINT_LIGHT = iota
	DIR_LIGHT
	SPOT_LIGHT
	AMBIENT_LIGHT
)

type Light struct {
	Type      int
	Pos       y3d.Vec3
	Intensity float32
	Direction y3d.Vec3
	Color     y3d.Color
	SpecColor y3d.Color
}
