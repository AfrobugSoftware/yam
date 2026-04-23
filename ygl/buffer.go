package ygl

import (
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type VertBuffer struct {
	Buf       uint32
	Indx      uint32
	VertArray uint32
	IndxCount int32
}

type ComponentType int32

const (
	ComponentTypeFloat32 ComponentType = iota
	ComponentTypeByte
	ComponentTypeUint8
	ComponentTypeInt16
	ComponentTypeUint16
	ComponentTypeUint32
)

func (c ComponentType) GlComponentType() uint32 {
	switch c {
	case ComponentTypeFloat32:
		return gl.FLOAT
	case ComponentTypeByte:
		return gl.BYTE
	case ComponentTypeUint8:
		return gl.UNSIGNED_BYTE
	case ComponentTypeInt16:
		return gl.SHORT
	case ComponentTypeUint16:
		return gl.UNSIGNED_SHORT
	case ComponentTypeUint32:
		return gl.UNSIGNED_INT
	}
	return 0
}

func (c ComponentType) ByteSize() uintptr {
	switch c {
	case ComponentTypeFloat32:
		return unsafe.Sizeof(float32(0))
	case ComponentTypeUint16, ComponentTypeInt16:
		return unsafe.Sizeof(uint16(0))
	case ComponentTypeUint32:
		return unsafe.Sizeof(uint32(0))
	case ComponentTypeUint8, ComponentTypeByte:
		return 1
	}
	return 0
}

type DataFormat struct {
	Count         int32         // num of components of the vertex
	Stride        int32         //size of each vertex in bytes
	Offset        int32         // offset into the packed slice where the vertext starts in byted
	ComponentType ComponentType //the component type
}

func (b VertBuffer) SetActive() {
	gl.BindVertexArray(b.VertArray)
}

func (b VertBuffer) DrawBuffer() {
	gl.DrawElements(gl.TRIANGLES, b.IndxCount, gl.UNSIGNED_SHORT, nil)
}

func CreateVextexBuffer(data []byte, indx []uint16, formats []DataFormat) VertBuffer {
	var idx, vertArray, buf uint32
	gl.GenVertexArrays(1, &vertArray)
	gl.BindVertexArray(vertArray)
	gl.GenBuffers(1, &buf)

	//gl.CreateBuffers(1, &buf)
	//gl.NamedBufferStorageEXT(buf, len(data), gl.Ptr(&data[0]), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ARRAY_BUFFER, buf)
	gl.BufferData(gl.ARRAY_BUFFER, len(data), gl.Ptr(&data[0]), gl.STATIC_DRAW)

	for i, f := range formats {
		gl.EnableVertexAttribArray(uint32(i))
		gl.VertexAttribPointer(uint32(i), f.Count, f.ComponentType.GlComponentType(), false, f.Stride, gl.Ptr(uintptr(f.Offset)))
	}

	gl.GenBuffers(1, &idx)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, idx)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, int(uintptr(len(indx))*unsafe.Sizeof(uint16(0))), gl.Ptr(&indx[0]), gl.STATIC_DRAW)

	return VertBuffer{
		Buf:       buf,
		Indx:      idx,
		VertArray: vertArray,
		IndxCount: int32(len(indx)),
	}
}

func DestroyVertexBuffer(buffer VertBuffer) {
	gl.DeleteBuffers(1, &buffer.Buf)
	gl.DeleteBuffers(1, &buffer.Indx)
	gl.DeleteVertexArrays(1, &buffer.VertArray)
}
