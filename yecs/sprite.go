package yecs

import gl "github.com/chsc/gogl/gl33"

type Sprite struct {
	Buffer         string
	Program        string
	Textures       string
	CurTexture     int
	Culled         bool
	AssignUniforms func(e EntityId, program gl.Uint) error
}
