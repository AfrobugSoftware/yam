package ycore

import (
	"bytes"
	"fmt"
	"unsafe"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	NO_STATICBUF = -1
)

type StaticBuffer struct {
	Id               int
	Vao              uint32
	VertexVbo        uint32
	IndexVbo         uint32
	DrawCommandBo    uint32
	DrawIndexVbo     uint32
	WorldMatrixSSBO  uint32 //should I change this to an attribute ?
	MaterialUBO      uint32
	NumVertics       int32
	NumIndices       int32
	NumDrawCommands  int32
	NumOfInstnaces   int32
	Stride           int32
	SkinId           int
	IdxType          uint32
	Format           []VertexFormat
	VManager         *VertexCacheManager
	SkeletalAnimator *SkeletalAnimator
}

func CheckError() {
	ierr := gl.GetError()
	if ierr != gl.NO_ERROR {
		fmt.Printf("GL error: 0x%x\n", ierr)
	}
}

func NewStaticBuffer(
	vertexCacheManager *VertexCacheManager,
	dataV, dataI *bytes.Buffer,
	command []DrawCommand,
	instanceCount int,
	skinId int,
	stride int32,
	format []VertexFormat,
	idxType uint32,
	id int,
) *StaticBuffer {
	if len(command) == 0 || len(format) == 0 {
		return nil
	}
	var vao, vvbo, ivbo, dcbo, divbo, mssbo, mubo uint32
	gl.CreateVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.CreateBuffers(1, &vvbo)
	gl.NamedBufferStorage(vvbo, dataV.Len(), gl.Ptr(dataV.Bytes()), 0)
	for i, f := range format {
		gl.VertexArrayAttribBinding(vao, uint32(i), 0)
		gl.VertexArrayAttribFormat(vao, uint32(i), f.ComponentSize, f.Type, false, f.RelativeOffset)
		gl.EnableVertexArrayAttrib(vao, uint32(i))
	}
	gl.VertexArrayVertexBuffer(vao, 0, vvbo, 0, stride)

	gl.CreateBuffers(1, &ivbo)
	gl.NamedBufferStorage(ivbo, dataI.Len(), gl.Ptr(dataI.Bytes()), 0)
	gl.VertexArrayElementBuffer(vao, ivbo)

	//create draw indices
	di := make([]uint32, instanceCount)
	for i := range instanceCount {
		di[i] = uint32(i)
	}

	gl.CreateBuffers(1, &divbo)
	gl.NamedBufferStorage(divbo, int(instanceCount*int(unsafe.Sizeof(uint32(0)))), gl.Ptr(di), 0)
	gl.VertexArrayAttribBinding(vao, 10, 10)
	gl.VertexArrayVertexBuffer(vao, 10, divbo, 0, int32(unsafe.Sizeof(uint32(0))))
	gl.VertexArrayAttribIFormat(vao, 10, 1, gl.UNSIGNED_INT, 0) //relative offset should it be zero
	gl.VertexArrayVertexAttribDivisorEXT(vao, 10, 1)
	gl.EnableVertexArrayAttrib(vao, 10)

	gl.CreateBuffers(1, &dcbo)
	gl.NamedBufferStorage(dcbo, int(unsafe.Sizeof(DrawCommand{})*uintptr(len(command))), gl.Ptr(command), 0)

	//create the world matrix ssbo
	gl.CreateBuffers(1, &mssbo)
	gl.NamedBufferStorage(mssbo, int(unsafe.Sizeof(y3d.Mat4{})*uintptr(instanceCount)),
		nil, gl.MAP_WRITE_BIT)

	//load material
	gl.CreateBuffers(1, &mubo)
	gl.NamedBufferStorage(mubo, int(unsafe.Sizeof([16]float32{})),
		nil,
		gl.MAP_WRITE_BIT)

	return &StaticBuffer{
		Id:              id,
		Vao:             vao,
		VertexVbo:       vvbo,
		IndexVbo:        ivbo,
		DrawCommandBo:   dcbo,
		DrawIndexVbo:    divbo,
		WorldMatrixSSBO: mssbo,
		MaterialUBO:     mubo,
		NumVertics:      (int32(dataV.Len() / int(stride))),
		NumIndices:      (int32(dataI.Len() / 4)),
		NumDrawCommands: int32(len(command)),
		NumOfInstnaces:  int32(instanceCount),
		SkinId:          skinId,
		Format:          format,
		Stride:          stride,
		IdxType:         idxType,
		VManager:        vertexCacheManager,
	}
}

func (s *StaticBuffer) Render(world []y3d.Mat4) {
	if s.VManager.ActiveSkin != s.SkinId && s.SkinId != -1 {
		skin, err := s.VManager.RenderManager.SkinManager.GetSkin(s.SkinId)
		if err != nil {
			return
		}
		material, err := s.VManager.RenderManager.SkinManager.GetMaterial(skin.Material)
		if err != nil {
			return
		}
		mPtr := gl.MapNamedBufferRange(
			s.MaterialUBO,
			0,
			int(unsafe.Sizeof([16]float32{})),
			gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_BUFFER_BIT)
		if mPtr == nil {
			CheckError()
			panic("cannot set material for static buffer")
		}
		m := unsafe.Slice((*float32)(mPtr), 16)
		m[0] = material.Diffuse.X
		m[1] = material.Diffuse.Y
		m[2] = material.Diffuse.Z
		m[3] = 0.0
		m[4] = material.Ambient.X
		m[5] = material.Ambient.Y
		m[6] = material.Ambient.Z
		m[7] = 0.0
		m[8] = material.Specular.X
		m[9] = material.Specular.Y
		m[10] = material.Specular.Z
		m[11] = 0.0
		m[12] = material.Emissive.X
		m[13] = material.Emissive.Y
		m[14] = material.Emissive.Z
		m[15] = material.Shininess
		gl.UnmapNamedBuffer(s.MaterialUBO)

		for i, t := range skin.Texture {
			if t == EmptyTexture {
				break
			}
			tex, ok := s.VManager.RenderManager.SkinManager.Textures[t]
			if ok {
				gl.BindTextureUnit(uint32(i), tex.Handle)
			}
		}
		s.VManager.ActiveSkin = s.SkinId
	}
	ptr := gl.MapNamedBufferRange(
		s.WorldMatrixSSBO,
		0,
		int(len(world)*int(unsafe.Sizeof(y3d.Mat4{}))),
		gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_BUFFER_BIT,
	)
	if ptr == nil {
		CheckError()
		panic("cannot load world matrix")
	}
	slice := unsafe.Slice((*y3d.Mat4)(ptr), len(world))
	copy(slice, world)
	if !gl.UnmapNamedBuffer(s.WorldMatrixSSBO) {
		CheckError()
		panic("cannot release world matrix buffer")
	}

	gl.BindVertexArray(s.Vao)
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, WORLD_MATRIX_BINDING, s.WorldMatrixSSBO)
	gl.BindBufferBase(gl.UNIFORM_BUFFER, MATERIAL_UBO_BINDING, s.MaterialUBO)
	if s.SkeletalAnimator != nil {
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, JOINTS_SSBO_BINDING, s.SkeletalAnimator.JointBufferObject)
	}
	gl.BindBuffer(gl.DRAW_INDIRECT_BUFFER, s.DrawCommandBo)
	switch s.VManager.RenderManager.DrawMode {
	case gl.TRIANGLES, gl.LINES:
		gl.MultiDrawElementsIndirect(s.VManager.RenderManager.DrawMode,
			s.IdxType,
			nil,
			s.NumDrawCommands,
			int32(unsafe.Sizeof(DrawCommand{})))
	case gl.POINTS:
		gl.MultiDrawArraysIndirect(s.VManager.RenderManager.DrawMode,
			nil,
			int32(s.NumDrawCommands),
			int32(unsafe.Sizeof(DrawCommand{})))
	}
}

func (s *StaticBuffer) Destory() {
	gl.DeleteBuffers(1, &s.IndexVbo)
	gl.DeleteBuffers(1, &s.VertexVbo)
	gl.DeleteBuffers(1, &s.DrawCommandBo)
	gl.DeleteBuffers(1, &s.DrawIndexVbo)
	gl.DeleteBuffers(1, &s.WorldMatrixSSBO)
	gl.DeleteBuffers(1, &s.MaterialUBO)
	gl.DeleteVertexArrays(1, &s.Vao)
}
