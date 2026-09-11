package ymanager

import (
	"bytes"
	"errors"
	"unsafe"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type VertexFormat struct {
	ComponentSize  int32
	Type           uint32
	RelativeOffset uint32
}

type DrawCommand struct {
	VertexCount   uint32
	InstanceCount uint32
	FirstIndex    uint32
	BaseVertex    uint32
	BaseInstance  uint32
}

type VertexCache struct {
	Vao             uint32
	VertexVbo       uint32
	IndexVbo        uint32
	DrawCommandBo   uint32
	DrawIndexVbo    uint32
	WorldMatrixSSBO uint32
	MaxVertices     int32
	MaxIndices      int32
	MaxDrawCommands int32
	NumVertics      int32
	NumIndices      int32
	NumDrawCommands int32
	Stride          int32
	SkinId          int
	Id              int
	Format          []VertexFormat
	SkinManager     *SkinManager
}

func NewVertexCache(
	skinmanager *SkinManager,
	maxVertex, maxIndex, maxDraws int32,
	stride int32,
	skinId int,
	id int,
	format []VertexFormat,
) *VertexCache {
	var vao, vvbo, ivbo, dcbo, divbo, mssbo uint32
	gl.CreateVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.CreateBuffers(1, &vvbo)
	gl.NamedBufferStorage(vvbo, int(maxVertex*stride), nil, gl.DYNAMIC_STORAGE_BIT)
	for i, f := range format {
		gl.VertexArrayAttribBinding(vao, uint32(i), 0)
		gl.VertexArrayAttribFormat(vao, uint32(i), f.ComponentSize, f.Type, false, f.RelativeOffset)
		gl.EnableVertexArrayAttrib(vao, uint32(i))
	}
	gl.VertexArrayVertexBuffer(vao, 0, vvbo, 0, stride)

	gl.CreateBuffers(1, &ivbo)
	gl.NamedBufferStorage(ivbo, int(maxIndex*4), nil, gl.DYNAMIC_STORAGE_BIT)
	gl.VertexArrayElementBuffer(vao, ivbo)

	gl.CreateBuffers(1, &divbo)
	gl.NamedBufferStorage(divbo, int(maxDraws*int32(unsafe.Sizeof(uint32(0)))), nil, gl.DYNAMIC_STORAGE_BIT)
	gl.VertexArrayAttribBinding(vao, 10, 10)
	gl.VertexArrayVertexBuffer(vao, 10, divbo, 0, int32(unsafe.Sizeof(uint32(0))))
	gl.VertexArrayAttribIFormat(vao, 10, 1, gl.UNSIGNED_INT, 0)
	gl.VertexArrayVertexAttribDivisorEXT(vao, 10, 1)
	gl.EnableVertexArrayAttrib(vao, 10)

	gl.CreateBuffers(1, &dcbo)
	gl.NamedBufferStorage(dcbo, int(unsafe.Sizeof(DrawCommand{})*uintptr(maxDraws)), nil, gl.DYNAMIC_STORAGE_BIT)

	//create the world matrix ssbo
	gl.CreateBuffers(1, &mssbo)
	gl.NamedBufferStorage(mssbo, int(unsafe.Sizeof(y3d.Mat4{})*uintptr(maxDraws)), nil, gl.DYNAMIC_STORAGE_BIT)

	return &VertexCache{
		Vao:             vao,
		VertexVbo:       vvbo,
		IndexVbo:        ivbo,
		DrawCommandBo:   dcbo,
		DrawIndexVbo:    divbo,
		WorldMatrixSSBO: mssbo,
		MaxVertices:     maxVertex,
		MaxIndices:      maxIndex,
		MaxDrawCommands: maxDraws,
		Stride:          stride,
		Id:              id,
		SkinId:          skinId,
		SkinManager:     skinmanager,
		Format:          format,
	}
}

func (v *VertexCache) Add(command DrawCommand, world y3d.Mat4, dataV, dataI *bytes.Buffer) error {
	if (v.Stride*v.MaxVertices) >= int32(dataV.Len()) ||
		(v.MaxIndices*4) >= int32(dataI.Len()) ||
		v.NumDrawCommands >= v.MaxDrawCommands {
		return errors.New("vertex cache full")
	}
	//vertex data
	d := dataV.Bytes()
	gl.NamedBufferSubData(v.VertexVbo,
		int(v.Stride*v.NumVertics), dataV.Len(), gl.Ptr(&d[0]))
	v.NumVertics += int32(dataV.Len() / int(v.Stride))
	//index data
	d = dataI.Bytes()
	gl.NamedBufferSubData(v.IndexVbo, int(4*v.NumIndices), dataI.Len(), gl.Ptr(&d[0]))
	command.BaseVertex = uint32(v.NumIndices)
	v.NumIndices += int32(dataI.Len() / 4)
	//draw commands
	gl.NamedBufferSubData(v.DrawCommandBo, int(unsafe.Sizeof(command)*uintptr(v.NumDrawCommands)),
		int(unsafe.Sizeof(command)), gl.Ptr(&command))
	gl.NamedBufferSubData(v.DrawIndexVbo, int(4*v.NumDrawCommands), int(unsafe.Sizeof(uint32(0))),
		gl.Ptr(&v.NumDrawCommands))
	//world transforms
	gl.NamedBufferSubData(v.WorldMatrixSSBO, int(unsafe.Sizeof(world)*uintptr(v.NumDrawCommands)),
		int(unsafe.Sizeof(world)),
		gl.Ptr(&world))

	v.NumDrawCommands += 1

	return nil
}
func (v *VertexCache) IsEmpty() bool {
	return v.NumVertics == 0
}

func (v *VertexCache) Flush() {
	//having the flush draw instead of the geometry object is somewhat confusing
	//not sure I understand that
}
func (v *VertexCache) Clear() {
	c := uint8(0x00)
	gl.BindBuffer(gl.ARRAY_BUFFER, v.VertexVbo)
	gl.ClearBufferData(gl.ARRAY_BUFFER, gl.R8, gl.R8, gl.UNSIGNED_BYTE, gl.Ptr(&c))
	gl.BindBuffer(gl.ARRAY_BUFFER, v.DrawIndexVbo)
	gl.ClearBufferData(gl.ARRAY_BUFFER, gl.R8, gl.R8, gl.UNSIGNED_BYTE, gl.Ptr(&c))
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, v.IndexVbo)
	gl.ClearBufferData(gl.ELEMENT_ARRAY_BUFFER, gl.R8, gl.R8, gl.UNSIGNED_BYTE, gl.Ptr(&c))

	gl.BindBuffer(gl.DRAW_INDIRECT_BUFFER, v.DrawCommandBo)
	gl.ClearBufferData(gl.DRAW_INDIRECT_BUFFER, gl.R8, gl.R8, gl.UNSIGNED_BYTE, gl.Ptr(&c))

	gl.BindBuffer(gl.SHADER_STORAGE_BUFFER, v.WorldMatrixSSBO)
	gl.ClearBufferData(gl.SHADER_STORAGE_BUFFER, gl.R8, gl.R8, gl.UNSIGNED_BYTE, gl.Ptr(&c))

	v.NumIndices = 0
	v.NumVertics = 0
}
func (v *VertexCache) SetSkin(skin int) {
	if !v.IsEmpty() {
		v.Flush()
	}
	v.SkinId = skin

	//set active cache
}

func (v *VertexCache) Destroy() {
	gl.DeleteBuffers(1, &v.IndexVbo)
	gl.DeleteBuffers(1, &v.VertexVbo)
	gl.DeleteBuffers(1, &v.DrawCommandBo)
	gl.DeleteBuffers(1, &v.DrawIndexVbo)
	gl.DeleteBuffers(1, &v.WorldMatrixSSBO)
	gl.DeleteVertexArrays(1, &v.Vao)
}
