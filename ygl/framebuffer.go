package ygl

import gl "github.com/chsc/gogl/gl33"

type FrameBuffer struct {
	Width, Height int
	Fbo           gl.Uint
	Color         gl.Uint
	Depth         gl.Uint
}
