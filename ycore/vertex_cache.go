package ycore

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type PerObjectRendeState struct {
	Flag  uint32
	Alpha float32
}

type VertexFormat struct {
	ComponentSize  int32
	Type           uint32
	RelativeOffset uint32
	Normalized     bool
}

type DrawCommand struct {
	Count         uint32
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
	RenderStateSSBO uint32
	MaxVertices     int32
	MaxIndices      int32
	MaxDrawCommands int32
	MaxInstances    int32
	NumVertics      int32
	NumIndices      int32
	NumDrawCommands int32
	NumOfInstances  int32
	Stride          int32
	SkinId          int
	Id              int
	Format          []VertexFormat
	SkinManager     *SkinManager
	VManager        *VertexCacheManager
}

func NewVertexCache(
	skinmanager *SkinManager,
	vertexManger *VertexCacheManager,
	maxVertex, maxIndex, maxDraws, maxInstances int32,
	stride int32,
	skinId int,
	id int,
	format []VertexFormat,
) *VertexCache {
	var vao, vvbo, ivbo, dcbo, divbo, mssbo, rfssobo uint32
	gl.CreateVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.CreateBuffers(1, &vvbo)
	gl.NamedBufferStorage(vvbo, int(maxVertex*stride), nil, gl.DYNAMIC_STORAGE_BIT)
	UpdateGPUMem(uint64(maxVertex * stride))
	for i, f := range format {
		gl.VertexArrayAttribBinding(vao, uint32(i), VERTEX_ATTRIBUTE_BINDING)
		switch f.Type {
		case gl.UNSIGNED_INT, gl.UNSIGNED_BYTE, gl.UNSIGNED_SHORT:
			gl.VertexArrayAttribIFormat(vao, uint32(i), f.ComponentSize, f.Type, f.RelativeOffset)
		default:
			gl.VertexArrayAttribFormat(vao, uint32(i), f.ComponentSize, f.Type, f.Normalized, f.RelativeOffset)
		}
		gl.EnableVertexArrayAttrib(vao, uint32(i))
	}
	gl.VertexArrayVertexBuffer(vao, VERTEX_ATTRIBUTE_BINDING, vvbo, 0, stride)

	gl.CreateBuffers(1, &ivbo)
	gl.NamedBufferStorage(ivbo, int(maxIndex*4), nil, gl.DYNAMIC_STORAGE_BIT)
	UpdateGPUMem(uint64(maxIndex * 4))
	gl.VertexArrayElementBuffer(vao, ivbo)

	gl.CreateBuffers(1, &divbo)
	gl.NamedBufferStorage(divbo, int(maxInstances*int32(unsafe.Sizeof(uint32(0)))), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	UpdateGPUMem(uint64(maxInstances * int32(unsafe.Sizeof(uint32(0)))))
	gl.VertexArrayAttribBinding(vao, 10, DRAW_INDEX_BINDING)
	gl.VertexArrayVertexBuffer(vao, DRAW_INDEX_BINDING, divbo, 0, int32(unsafe.Sizeof(uint32(0))))
	gl.VertexArrayAttribIFormat(vao, 10, 1, gl.UNSIGNED_INT, 0)
	gl.VertexArrayVertexAttribDivisorEXT(vao, 10, 1)
	gl.EnableVertexArrayAttrib(vao, 10)

	gl.CreateBuffers(1, &dcbo)
	gl.NamedBufferStorage(dcbo, int(unsafe.Sizeof(DrawCommand{})*uintptr(maxDraws)), nil, gl.DYNAMIC_STORAGE_BIT)
	UpdateGPUMem(uint64(unsafe.Sizeof(DrawCommand{}) * uintptr(maxDraws)))
	//create the world matrix ssbo
	gl.CreateBuffers(1, &mssbo)
	gl.NamedBufferStorage(mssbo, int((unsafe.Sizeof(y3d.Mat4{})*2)*uintptr(maxInstances)), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	UpdateGPUMem(uint64(unsafe.Sizeof(y3d.Mat4{}) * uintptr(maxInstances)))
	gl.CreateBuffers(1, &rfssobo)
	gl.NamedBufferStorage(rfssobo, int(unsafe.Sizeof(PerObjectRendeState{})*uintptr(maxInstances)), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	UpdateGPUMem(uint64(unsafe.Sizeof(PerObjectRendeState{}) * uintptr(maxInstances)))
	return &VertexCache{
		Vao:             vao,
		VertexVbo:       vvbo,
		IndexVbo:        ivbo,
		DrawCommandBo:   dcbo,
		DrawIndexVbo:    divbo,
		WorldMatrixSSBO: mssbo,
		RenderStateSSBO: rfssobo,
		MaxVertices:     maxVertex,
		MaxIndices:      maxIndex,
		MaxDrawCommands: maxDraws,
		MaxInstances:    maxInstances,
		Stride:          stride,
		Id:              id,
		SkinId:          skinId,
		SkinManager:     skinmanager,
		VManager:        vertexManger,
		Format:          format,
	}
}
func (v *VertexCache) IsFull(size int) bool {
	n := size / int(v.Stride)
	return v.NumVertics+int32(n) >= v.MaxVertices
}
func (v *VertexCache) AddPolygon(indices []uint32,
	vertices []y3d.PVertex) error {
	if int32(len(vertices)*12) >= (v.Stride*v.MaxVertices) ||
		int32(len(indices)*4) >= (v.MaxIndices*4) {
		return errors.New("cannot add data, please check parameters sizes")
	}
	if v.Stride != 36 {
		return errors.New("cannot write to this cache, stride size not 36")
	}
	for i, x := range indices {
		indices[i] = x + uint32(v.NumVertics)
	}
	asBytes := func(v reflect.Value, count int) []byte {
		addr := unsafe.Pointer(v.Index(0).UnsafeAddr())
		return unsafe.Slice((*byte)(addr), count)
	}
	size := len(vertices) * int(unsafe.Sizeof(y3d.PVertex{}))
	vb := asBytes(reflect.ValueOf(vertices), size)
	ib := asBytes(reflect.ValueOf(indices), len(indices)*4)
	count := len(vertices)
	gl.NamedBufferSubData(v.VertexVbo,
		int(v.Stride*v.NumVertics),
		size,
		gl.Ptr(vb))
	v.NumVertics += int32(count)
	gl.NamedBufferSubData(v.IndexVbo,
		int(4*v.NumIndices),
		len(ib),
		gl.Ptr(ib))
	v.NumIndices += int32(len(indices))
	return CheckError()
}

func (v *VertexCache) Add(command *DrawCommand,
	dataV, dataI *bytes.Buffer) error {
	if dataV == nil || dataI == nil {
		return errors.New("cannot load empty vertex or index buffer")
	}
	if int32(dataV.Len()) >= (v.Stride*v.MaxVertices) ||
		int32(dataI.Len()) >= (v.MaxIndices*4) {
		return errors.New("cannot add data, please check parameters")
	}
	gl.NamedBufferSubData(v.VertexVbo,
		int(v.Stride*v.NumVertics),
		dataV.Len(),
		gl.Ptr(dataV.Bytes()))
	command.BaseVertex = uint32(v.NumVertics)
	v.NumVertics += int32(dataV.Len() / int(v.Stride))
	gl.NamedBufferSubData(v.IndexVbo,
		int(4*v.NumIndices),
		dataI.Len(),
		gl.Ptr(dataI.Bytes()))
	v.NumIndices += int32(dataI.Len() / 4)
	command.Count = uint32(dataI.Len() / 4)
	command.InstanceCount = 1 //always set to one
	return CheckError()
}
func (v *VertexCache) IsEmpty() bool {
	return v.NumVertics == 0 || v.NumDrawCommands == 0
}

func (v *VertexCache) AddInstances(command *DrawCommand) error {
	if v.NumOfInstances >= v.MaxInstances {
		return errors.New("no space for new instances")
	}
	command.BaseInstance = uint32(v.NumOfInstances)
	instanceCount := int(command.InstanceCount)
	ptr := gl.MapNamedBufferRange(
		v.DrawIndexVbo,
		int(unsafe.Sizeof(uint32(0)))*int(command.BaseInstance),
		int(unsafe.Sizeof(uint32(0)))*instanceCount,
		gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_RANGE_BIT,
	)
	if ptr == nil {
		return CheckError()
	}
	id := unsafe.Slice((*uint32)(ptr), instanceCount)
	for i := range instanceCount {
		id[i] = uint32(i + int(command.BaseInstance))
	}
	if !gl.UnmapNamedBuffer(v.DrawIndexVbo) {
		return CheckError()
	}
	v.NumOfInstances += int32(instanceCount)
	return CheckError()
}

func (v *VertexCache) AddDrawCommand(command DrawCommand) error {
	if v.NumDrawCommands >= v.MaxDrawCommands {
		return errors.New("no space for command")
	}
	gl.NamedBufferSubData(v.DrawCommandBo, int(unsafe.Sizeof(command)*uintptr(v.NumDrawCommands)),
		int(unsafe.Sizeof(command)), unsafe.Pointer(&command))
	v.NumDrawCommands++
	return CheckError()
}

func (v *VertexCache) AddRenderState(baseInstance int, flag ...PerObjectRendeState) error {
	if baseInstance+len(flag) >= int(v.MaxInstances) {
		return errors.New("invalid matrix instance range")
	}
	ptr := gl.MapNamedBufferRange(
		v.RenderStateSSBO,
		int(unsafe.Sizeof(PerObjectRendeState{}))*int(baseInstance),
		int(len(flag)*int(unsafe.Sizeof(PerObjectRendeState{}))),
		gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_RANGE_BIT,
	)
	if ptr == nil {
		CheckError()
		return errors.New("cannot load world matrix")
	}
	slice := unsafe.Slice((*PerObjectRendeState)(ptr), len(flag))
	copy(slice, flag)
	if !gl.UnmapNamedBuffer(v.RenderStateSSBO) {
		CheckError()
		return errors.New("cannot release world matrix buffer")
	}
	return CheckError()
}

func (v *VertexCache) LoadMatrix(baseInstance int, world ...y3d.Mat4) error {
	if baseInstance+len(world) >= int(v.MaxInstances) {
		return errors.New("invalid matrix instance range")
	}
	ptr := gl.MapNamedBufferRange(
		v.WorldMatrixSSBO,
		(int(unsafe.Sizeof(y3d.Mat4{}))*2)*int(baseInstance),
		int(len(world)*int(unsafe.Sizeof(y3d.Mat4{}))),
		gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_RANGE_BIT,
	)
	if ptr == nil {
		CheckError()
		return errors.New("cannot load world matrix")
	}
	type outObject struct {
		world  y3d.Mat4
		normal y3d.Mat4
	}
	out := make([]outObject, len(world))
	for i := range world {
		out[i].world = world[i]
		world[i][12] = 0
		world[i][13] = 0
		world[i][14] = 0
		out[i].normal = world[i]
		(&out[i].normal).Invert()
		(&out[i].normal).Transpose()
	}
	slice := unsafe.Slice((*outObject)(ptr), len(world))
	copy(slice, out)
	if !gl.UnmapNamedBuffer(v.WorldMatrixSSBO) {
		CheckError()
		return errors.New("cannot release world matrix buffer")
	}
	return CheckError()
}

func (v *VertexCache) Render() {
	if v.IsEmpty() {
		return
	}
	if v.NumDrawCommands != 0 {
		//setup skin
		if v.VManager.ActiveSkin != v.SkinId && v.SkinId != -1 {
			v.VManager.RenderManager.SkinManager.BindSkin(v.SkinId)
			v.VManager.ActiveSkin = v.SkinId
		}
		//how to handle framebuffer operations ???
		gl.BindVertexArray(v.Vao)
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, WORLD_MATRIX_BINDING, v.WorldMatrixSSBO)
		gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, RENDER_FLAG_BINDING, v.RenderStateSSBO)
		gl.BindBuffer(gl.DRAW_INDIRECT_BUFFER, v.DrawCommandBo)
		//who should have draw mode???
		switch v.VManager.RenderManager.DrawMode {
		case gl.TRIANGLES, gl.LINES, gl.LINE_STRIP:
			gl.MultiDrawElementsIndirect(v.VManager.RenderManager.DrawMode,
				gl.UNSIGNED_INT, nil, v.NumDrawCommands, int32(unsafe.Sizeof(DrawCommand{})))
		case gl.POINTS:
			gl.MultiDrawArraysIndirect(v.VManager.RenderManager.DrawMode, nil, int32(v.NumDrawCommands),
				int32(unsafe.Sizeof(DrawCommand{})))
		}
		//reset draw command buffer
		v.NumDrawCommands = 0
		v.NumOfInstances = 0
	}
}

func (v *VertexCache) Reset() {
	v.NumIndices = 0
	v.NumVertics = 0
	v.NumDrawCommands = 0
	v.NumOfInstances = 0
}

func (v *VertexCache) SetSkin(skin int) {
	if !v.IsEmpty() {
		v.Render()
		v.Reset()
	}
	v.SkinId = skin
}

func (v *VertexCache) Destroy() {
	gl.DeleteBuffers(1, &v.IndexVbo)
	gl.DeleteBuffers(1, &v.VertexVbo)
	gl.DeleteBuffers(1, &v.DrawCommandBo)
	gl.DeleteBuffers(1, &v.DrawIndexVbo)
	gl.DeleteBuffers(1, &v.WorldMatrixSSBO)
	gl.DeleteBuffers(1, &v.RenderStateSSBO)
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
