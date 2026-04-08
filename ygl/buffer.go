package ygl

import (
	"unsafe"

	gl "github.com/chsc/gogl/gl33"
)

type VertBuffer struct {
	Buf       gl.Uint
	Indx      gl.Uint
	VertArray gl.Uint
	IndxCount int
}

type DataFormat struct {
	Count  int // num of components of the vertex
	Stride int //size of each vertex in bytes
	Offset int // offset into the packed slice where the vertext starts
}

func (b VertBuffer) SetActive() {
	gl.BindVertexArray(b.VertArray)
}

func (b VertBuffer) DrawBuffer() {
	gl.DrawElements(gl.TRIANGLES, gl.Sizei(b.IndxCount), gl.UNSIGNED_SHORT, nil)
}

func CreateVextexBuffer(data []gl.Float, indx []uint16, formats []DataFormat) VertBuffer {
	var idx, vertArray, buf gl.Uint
	byteSize := unsafe.Sizeof(gl.Float(0))
	gl.GenVertexArrays(1, &vertArray)
	gl.BindVertexArray(vertArray)
	gl.GenBuffers(1, &buf)
	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(data)*int(byteSize)), gl.Pointer(&data[0]), gl.STATIC_DRAW)

	for i, f := range formats {
		gl.EnableVertexAttribArray(gl.Uint(i))
		gl.VertexAttribPointer(gl.Uint(i), gl.Int(f.Count), gl.FLOAT, gl.FALSE, gl.Sizei(f.Stride), gl.Pointer(byteSize*uintptr(f.Offset)))
	}

	gl.GenBuffers(1, &idx)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, idx)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(uintptr(len(indx))*unsafe.Sizeof(uint16(0))), gl.Pointer(&indx[0]), gl.STATIC_DRAW)

	return VertBuffer{
		Buf:       buf,
		Indx:      idx,
		VertArray: vertArray,
		IndxCount: len(indx),
	}
}

func DestroyVertexBuffer(buffer VertBuffer) {
	gl.DeleteBuffers(1, &buffer.Buf)
	gl.DeleteBuffers(1, &buffer.Indx)
	gl.DeleteVertexArrays(1, &buffer.VertArray)
}
