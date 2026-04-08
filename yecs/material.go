package yecs

import "yam/y3d"

type Material struct {
	Diffuse   y3d.Color
	Ambient   y3d.Color
	Specular  y3d.Color
	Shininess float32
}
