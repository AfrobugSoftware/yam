package yecs

import gl "github.com/chsc/gogl/gl33"

type Sprite struct {
	Buffer   string
	Program  string
	Textures string

	AssignUniforms func(program gl.Uint) error
}
