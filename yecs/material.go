package yecs

import "yam/y3d"

var (
	IdentityMaterial = Material{
		Diffuse:   y3d.Vec3{X: 1, Y: 1, Z: 1},
		Ambient:   y3d.Vec3{X: 1, Y: 1, Z: 1},
		Specular:  y3d.Vec3{X: 1, Y: 1, Z: 1},
		Shininess: 1.0,
	}
)

type Material struct {
	Diffuse   y3d.Vec3
	Ambient   y3d.Vec3
	Specular  y3d.Vec3
	Shininess float32
}

type MaterialSurface struct {
	Diffuse  uint32
	Specular uint32
	Normal   uint32
}
