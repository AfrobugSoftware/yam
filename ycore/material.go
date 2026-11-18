package ycore

import (
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

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

func (p *PbrMaterial) Close() {
	if p.Albeo != 0 {
		gl.DeleteTextures(1, &p.Albeo)
		p.Albeo = 0
	}
	if p.Specular != 0 {
		gl.DeleteTextures(1, &p.Specular)
		p.Specular = 0
	}
	if p.NormalMap != 0 {
		gl.DeleteTextures(1, &p.NormalMap)
		p.NormalMap = 0
	}
}

type PbrMaterialData struct {
	BaseColorFactor y3d.Vec4
	MetallicFactor  float32
	RoughnessFactor float32
	EmissiveFactor  y3d.Vec3
}
