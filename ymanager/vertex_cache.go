package ymanager

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
	Vao               uint32
	VertexVbo         uint32
	IndexVbo          uint32
	DrawCommandBo     uint32
	DrawIndexVbo      uint32
	WorldMatrixSSBO   uint32
	MaterialUBO       uint32
	MaxVertices       int32
	MaxIndices        int32
	MaxDrawCommands   int32
	MaxWorldMatrics   int32
	NumVertics        int32
	NumIndices        int32
	NumDrawCommands   int32
	NumOfWorldMatrics int32
	Stride            int32
	SkinId            int
	Id                int
	Format            []VertexFormat
	SkinManager       *SkinManager
	VManager          *VertexCacheManager
}

func NewVertexCache(
	skinmanager *SkinManager,
	maxVertex, maxIndex, maxDraws, maxMatrix int32,
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
	gl.NamedBufferStorage(mssbo, int(unsafe.Sizeof(y3d.Mat4{})*uintptr(maxMatrix)), nil, gl.DYNAMIC_STORAGE_BIT)

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
		MaxWorldMatrics: maxMatrix,
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

func (v *VertexCache) Add(command DrawCommand, world []y3d.Mat4, dataV, dataI *bytes.Buffer) error {
	if (v.Stride*v.MaxVertices) >= int32(dataV.Len()) ||
		(v.MaxIndices*4) >= int32(dataI.Len()) ||
		v.NumDrawCommands >= v.MaxDrawCommands {
		return errors.New("vertex cache full")
	}
	//vertex data
	d := dataV.Bytes()
	gl.NamedBufferSubData(v.VertexVbo,
		int(v.Stride*v.NumVertics), dataV.Len(), gl.Ptr(&d[0]))
	command.BaseVertex = uint32(v.NumVertics)
	v.NumVertics += int32(dataV.Len() / int(v.Stride))
	//index data
	d = dataI.Bytes()
	gl.NamedBufferSubData(v.IndexVbo, int(4*v.NumIndices), dataI.Len(), gl.Ptr(&d[0]))
	v.NumIndices += int32(dataI.Len() / 4)
	//draw commands
	command.BaseInstance = uint32(v.NumOfWorldMatrics)
	gl.NamedBufferSubData(v.DrawCommandBo, int(unsafe.Sizeof(command)*uintptr(v.NumDrawCommands)),
		int(unsafe.Sizeof(command)), gl.Ptr(&command))
	id := make([]uint32, len(world))
	for i := range world {
		id[i] = uint32(i + int(command.BaseInstance))
	}
	gl.NamedBufferSubData(
		v.DrawIndexVbo,
		int(4*v.NumOfWorldMatrics),
		int(4*len(id)),
		gl.Ptr(&id[0]))
	v.NumDrawCommands += 1

	//world transforms
	gl.NamedBufferSubData(v.WorldMatrixSSBO, int(unsafe.Sizeof(y3d.Mat4{})*uintptr(v.NumOfWorldMatrics)),
		int(unsafe.Sizeof(world)),
		gl.Ptr(&world[0]))
	v.NumOfWorldMatrics += int32(len(world))
	return nil
}
func (v *VertexCache) IsEmpty() bool {
	return v.NumVertics == 0
}

func (v *VertexCache) Flush() {
	//having the flush draw instead of the geometry object is somewhat confusing
	//not sure I understand that
	if v.NumDrawCommands != 0 {
		//setup skin
		if v.VManager.ActiveCache != v.Id {
			if v.VManager.ActiveSkin != v.SkinId {
				skin, err := v.VManager.RenderManager.SkinManager.GetSkin(v.SkinId)
				if err != nil {
					return
				}
				material, err := v.VManager.RenderManager.SkinManager.GetMaterial(skin.Material)
				if err != nil {
					return
				}
				gl.BindBuffer(gl.UNIFORM_BUFFER, v.MaterialUBO)
				mPtr := gl.MapBuffer(gl.UNIFORM_BUFFER, gl.WRITE_ONLY)
				if mPtr != nil {
					panic(fmt.Sprintf("material buffer not set up for: %d", v.Id))
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
				gl.UnmapBuffer(gl.UNIFORM_BUFFER)

				for _, i := range skin.Texture {
					if i == EmptyTexture {
						break
					}
					tex, ok := v.VManager.RenderManager.SkinManager.Textures[i]
					if ok {
						gl.BindTextureUnit(uint32(i), tex.Handle)
					}
				}
				v.VManager.ActiveSkin = v.SkinId
			}
			//how to handle render states ???

			v.VManager.ActiveCache = v.Id
			gl.BindVertexArray(v.Vao)
			gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, 10, v.WorldMatrixSSBO)
			gl.BindBufferBase(gl.UNIFORM_BUFFER, 16, v.MaterialUBO)
			switch v.VManager.RenderManager.DrawMode {
			case gl.TRIANGLES, gl.LINES:
				gl.MultiDrawElementsIndirect(v.VManager.RenderManager.DrawMode,
					gl.UNSIGNED_INT, nil, v.NumDrawCommands, int32(unsafe.Sizeof(DrawCommand{})))
			case gl.POINTS:
				gl.MultiDrawArraysIndirect(v.VManager.RenderManager.DrawMode, nil, int32(v.NumDrawCommands),
					int32(unsafe.Sizeof(DrawCommand{})))
			}
			v.NumDrawCommands = 0
			v.NumIndices = 0
			v.NumVertics = 0
		}
	}
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
	v.VManager.ActiveCache = INVALID_CACHE
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
