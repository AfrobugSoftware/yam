package ycore

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

type PbrMaterial struct {
	Albeo     uint32
	Specular  uint32
	NormalMap uint32
	PbrParam  uint32 //index into the pbr material data buffer
}

type PbrMaterialData struct {
	BaseColorFactor y3d.Vec4
	MetallicFactor  float32
	RoughnessFactor float32
	EmissiveFactor  y3d.Vec3
}
