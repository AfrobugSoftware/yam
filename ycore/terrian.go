package ycore

import (
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
)

// for test
type Terrian struct {
	VertexArray  uint32
	VertexBuffer uint32
	IndexBuffer  uint32
	Textures     []uint32
	Program      uint32
}

func NewTerrian(maxVertices, maxIndex, stride int) *Terrian {
	var vao, vvbo, ivbo uint32
	gl.CreateVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	buf := []float32{
		-1.00, 0, -1.00,
		-1.00, 0, 1.00,
		1.00, 0, -1.00,
		1.00, 0, 1.00,
	}
	gl.CreateBuffers(1, &vvbo)
	gl.NamedBufferStorage(vvbo, int(unsafe.Sizeof(float32(0)))*len(buf),
		gl.Ptr(buf),
		0)

	gl.CreateBuffers(1, &ivbo)
	gl.NamedBufferStorage(ivbo, int(maxIndex*4), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT|gl.MAP_READ_BIT)
	gl.VertexArrayElementBuffer(vao, ivbo)
	return &Terrian{
		VertexArray:  vao,
		VertexBuffer: vvbo,
		IndexBuffer:  ivbo,
	}
}
