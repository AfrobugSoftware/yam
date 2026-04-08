package yecs

import gl "github.com/chsc/gogl/gl33"

type Spatial struct {
	Buffer         string
	Program        string
	Textures       string
	CurTexture     int
	AssignUniforms func(e EntityId, w *World, cam *Camera, program gl.Uint) error
}
