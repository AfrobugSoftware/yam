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
	Diffuse   y3d.Vec3
	Ambient   y3d.Vec3
	Specular  y3d.Vec3
}
