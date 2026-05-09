package yecs

import "yam/y3d"

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
