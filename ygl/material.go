package ygl

import "yam/y3d"

var (
	IdentityMaterial = Material{
		Diffuse:  y3d.Vec4{X: 1, Y: 1, Z: 1, W: 0},
		Ambient:  y3d.Vec4{X: 1, Y: 1, Z: 1, W: 0},
		Specular: y3d.Vec4{X: 1, Y: 1, Z: 1, W: 0},
		Emissive: y3d.Vec4{X: 1, Y: 1, Z: 1, W: 16},
	}
)

type Material struct {
	Diffuse  y3d.Vec4
	Ambient  y3d.Vec4
	Specular y3d.Vec4
	Emissive y3d.Vec4
}
