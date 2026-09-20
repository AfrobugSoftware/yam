package ycore

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
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
	Vao              uint32
	VertexVbo        uint32
	IndexVbo         uint32
	DrawCommandBo    uint32
	DrawIndexVbo     uint32
	WorldMatrixSSBO  uint32
	MaterialUBO      uint32
	MaxVertices      int32
	MaxIndices       int32
	MaxDrawCommands  int32
	MaxInstances     int32
	NumVertics       int32
	NumIndices       int32
	NumDrawCommands  int32
	NumOfInstances   int32
	Stride           int32
	SkinId           int
	Id               int
	Format           []VertexFormat
	SkinManager      *SkinManager
	VManager         *VertexCacheManager
	SkeletalAnimator *SkeletalAnimator
}

func NewVertexCache(
	skinmanager *SkinManager,
	maxVertex, maxIndex, maxDraws, maxInstances int32,
	stride int32,
	skinId int,
	id int,
	format []VertexFormat,
) *VertexCache {
	var vao, vvbo, ivbo, dcbo, divbo, mssbo, mubo uint32
	gl.CreateVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.CreateBuffers(1, &vvbo)
	gl.NamedBufferStorage(vvbo, int(maxVertex*stride), nil, gl.DYNAMIC_STORAGE_BIT)
	for i, f := range format {
		gl.VertexArrayAttribBinding(vao, uint32(i), VERTEX_ATTRIBUTE_BINDING)
		switch f.Type {
		case gl.UNSIGNED_INT, gl.UNSIGNED_BYTE, gl.UNSIGNED_SHORT:
			gl.VertexArrayAttribIFormat(vao, uint32(i), f.ComponentSize, f.Type, f.RelativeOffset)
		default:
			gl.VertexArrayAttribFormat(vao, uint32(i), f.ComponentSize, f.Type, false, f.RelativeOffset)
		}
		gl.EnableVertexArrayAttrib(vao, uint32(i))
	}
	gl.VertexArrayVertexBuffer(vao, VERTEX_ATTRIBUTE_BINDING, vvbo, 0, stride)

	gl.CreateBuffers(1, &ivbo)
	gl.NamedBufferStorage(ivbo, int(maxIndex*4), nil, gl.DYNAMIC_STORAGE_BIT)
	gl.VertexArrayElementBuffer(vao, ivbo)

	gl.CreateBuffers(1, &divbo)
	gl.NamedBufferStorage(divbo, int(maxInstances*int32(unsafe.Sizeof(uint32(0)))), nil, gl.DYNAMIC_STORAGE_BIT)
	gl.VertexArrayAttribBinding(vao, 10, DRAW_INDEX_BINDING)
	gl.VertexArrayVertexBuffer(vao, DRAW_INDEX_BINDING, divbo, 0, int32(unsafe.Sizeof(uint32(0))))
	gl.VertexArrayAttribIFormat(vao, 10, 1, gl.UNSIGNED_INT, 0)
	gl.VertexArrayVertexAttribDivisorEXT(vao, 10, 1)
	gl.EnableVertexArrayAttrib(vao, 10)

	gl.CreateBuffers(1, &dcbo)
	gl.NamedBufferStorage(dcbo, int(unsafe.Sizeof(DrawCommand{})*uintptr(maxDraws)), nil, gl.DYNAMIC_STORAGE_BIT)

	//create the world matrix ssbo
	gl.CreateBuffers(1, &mssbo)
	gl.NamedBufferStorage(mssbo, int(unsafe.Sizeof(y3d.Mat4{})*uintptr(maxInstances)), nil, gl.DYNAMIC_STORAGE_BIT)

	gl.CreateBuffers(1, &mubo)
	gl.NamedBufferStorage(mubo, int(unsafe.Sizeof([16]float32{})), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)

	return &VertexCache{
		Vao:             vao,
		VertexVbo:       vvbo,
		IndexVbo:        ivbo,
		DrawCommandBo:   dcbo,
		DrawIndexVbo:    divbo,
		WorldMatrixSSBO: mssbo,
		MaterialUBO:     mubo,
		MaxVertices:     maxVertex,
		MaxIndices:      maxIndex,
		MaxDrawCommands: maxDraws,
		MaxInstances:    maxInstances,
		Stride:          stride,
		Id:              id,
		SkinId:          skinId,
		SkinManager:     skinmanager,
		Format:          format,
	}
}
func (v *VertexCache) IsFull(size int) bool {
	n := size / int(v.Stride)
	return v.NumVertics+int32(n) >= v.MaxVertices
}

func (v *VertexCache) Add(command DrawCommand,
	instanceCount int,
	world []y3d.Mat4,
	dataV, dataI *bytes.Buffer) error {
	if (v.Stride*v.MaxVertices) >= int32(dataV.Len()) ||
		(v.MaxIndices*4) >= int32(dataI.Len()) ||
		v.NumDrawCommands >= v.MaxDrawCommands ||
		v.NumOfInstances >= v.MaxInstances ||
		len(world) != instanceCount {
		return errors.New("cannot add data, please check parameters")
	}
	//vertex data
	gl.NamedBufferSubData(v.VertexVbo,
		int(v.Stride*v.NumVertics),
		dataV.Len(),
		gl.Ptr(dataV.Bytes()))
	command.BaseVertex = uint32(v.NumVertics)
	v.NumVertics += int32(dataV.Len() / int(v.Stride))
	//index data
	gl.NamedBufferSubData(v.IndexVbo,
		int(4*v.NumIndices),
		dataI.Len(),
		gl.Ptr(dataI.Bytes()))
	v.NumIndices += int32(dataI.Len() / 4)
	//draw commands
	command.BaseInstance = uint32(v.NumOfInstances)
	gl.NamedBufferSubData(v.DrawCommandBo, int(unsafe.Sizeof(command)*uintptr(v.NumDrawCommands)),
		int(unsafe.Sizeof(command)), gl.Ptr(&command))
	v.NumDrawCommands += 1

	id := make([]uint32, instanceCount)
	for i := range instanceCount {
		id[i] = uint32(i + int(command.BaseInstance))
	}
	gl.NamedBufferSubData(
		v.DrawIndexVbo,
		int(4*v.NumOfInstances),
		int(4*len(id)),
		gl.Ptr(id))

	//world transforms
	gl.NamedBufferSubData(v.WorldMatrixSSBO, int(unsafe.Sizeof(y3d.Mat4{})*uintptr(v.NumOfInstances)),
		int(unsafe.Sizeof(y3d.Mat4{}))*instanceCount,
		gl.Ptr(world))
	v.NumOfInstances += int32(instanceCount)

	return nil
}
func (v *VertexCache) IsEmpty() bool {
	return v.NumVertics == 0
}

func (v *VertexCache) Render() {
	if v.NumDrawCommands != 0 {
		//setup skin
		if v.VManager.ActiveSkin != v.SkinId && v.SkinId != -1 {
			skin, err := v.VManager.RenderManager.SkinManager.GetSkin(v.SkinId)
			if err != nil {
				return
			}
			material, err := v.VManager.RenderManager.SkinManager.GetMaterial(skin.Material)
			if err != nil {
				return
			}
			mPtr := gl.MapNamedBufferRange(
				v.MaterialUBO,
				0,
				4*16,
				gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_BUFFER_BIT)
			if mPtr != nil {
				panic("cannot set material for vertex cache")
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
			gl.UnmapNamedBuffer(v.MaterialUBO)

			for i, t := range skin.Texture {
				if t == EmptyTexture {
					break
				}
				tex, ok := v.VManager.RenderManager.SkinManager.Textures[t]
				if ok {
					gl.BindTextureUnit(uint32(i), tex.Handle)
				}
			}
			v.VManager.ActiveSkin = v.SkinId
		}
		//how to handle render states ???

		gl.BindVertexArray(v.Vao)
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, WORLD_MATRIX_BINDING, v.WorldMatrixSSBO)
		gl.BindBufferBase(gl.UNIFORM_BUFFER, MATERIAL_UBO_BINDING, v.MaterialUBO)
		if v.SkeletalAnimator != nil {
			gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, JOINTS_SSBO_BINDING, v.SkeletalAnimator.JointBufferObject)
		}
		gl.BindBuffer(gl.DRAW_INDIRECT_BUFFER, v.DrawCommandBo)
		switch v.VManager.RenderManager.DrawMode {
		case gl.TRIANGLES, gl.LINES, gl.LINE_STRIP:
			gl.MultiDrawElementsIndirect(v.VManager.RenderManager.DrawMode,
				gl.UNSIGNED_INT, nil, v.NumDrawCommands, int32(unsafe.Sizeof(DrawCommand{})))
		case gl.POINTS:
			gl.MultiDrawArraysIndirect(v.VManager.RenderManager.DrawMode, nil, int32(v.NumDrawCommands),
				int32(unsafe.Sizeof(DrawCommand{})))
		}
		v.Reset()
	}
}

func (v *VertexCache) Reset() {
	v.NumIndices = 0
	v.NumVertics = 0
	v.NumDrawCommands = 0
	v.NumOfInstances = 0
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
	v.NumDrawCommands = 0
	v.NumOfInstances = 0
}
func (v *VertexCache) SetSkin(skin int) {
	if !v.IsEmpty() {
		v.Render()
	}
	v.SkinId = skin
}

func (v *VertexCache) Destroy() {
	gl.DeleteBuffers(1, &v.IndexVbo)
	gl.DeleteBuffers(1, &v.VertexVbo)
	gl.DeleteBuffers(1, &v.DrawCommandBo)
	gl.DeleteBuffers(1, &v.DrawIndexVbo)
	gl.DeleteBuffers(1, &v.WorldMatrixSSBO)
	gl.DeleteBuffers(1, &v.MaterialUBO)
	gl.DeleteVertexArrays(1, &v.Vao)
}

func (v *VertexCache) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Cache Id: %d\n", v.Id)
	fmt.Fprintf(&b, "Vertex array object: %d\n", v.Vao)
	fmt.Fprintf(&b, "Vertices buffer: %d\n", v.VertexVbo)
	fmt.Fprintf(&b, "Index buffer : %d\n", v.IndexVbo)
	fmt.Fprintf(&b, "Draw command buffer: %d\n", v.DrawCommandBo)
	fmt.Fprintf(&b, "World matrix ssbo: %d\n", v.WorldMatrixSSBO)
	fmt.Fprintf(&b, "Material UBO: %d\n", v.MaterialUBO)
	fmt.Fprintf(&b, "Max vertices: %d\n", v.MaxVertices)
	fmt.Fprintf(&b, "Max Indices: %d\n", v.MaxIndices)
	fmt.Fprintf(&b, "Max Draw commands: %d\n", v.MaxDrawCommands)
	fmt.Fprintf(&b, "Number of vertices in buffer: %d\n", v.NumVertics)
	fmt.Fprintf(&b, "Number of indices in buffer: %d\n", v.NumIndices)
	fmt.Fprintf(&b, "Number of draw commands in buffer: %d\n", v.NumDrawCommands)
	fmt.Fprintf(&b, "Stride of vertex: %d\n", v.Stride)
	fmt.Fprintf(&b, "Skin id: %d\n", v.SkinId)

	b.WriteString("Vertex formats\n\n")
	for _, vf := range v.Format {
		fmt.Fprintf(&b, "\t\t%v\n", vf)
	}

	return b.String()
}

func (v *VertexFormat) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Component size: %d", v.ComponentSize)
	fmt.Fprintf(&b, "Relative offset: %d", v.RelativeOffset)
	fmt.Fprintf(&b, "Type: %d", v.Type)
	return b.String()
}
