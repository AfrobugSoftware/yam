package ycore

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"unsafe"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	INVALID_CACHE = -1
	MAX_CACHES    = 10
)

const (
	VP       = "pos"
	VPNT     = "pos.normal.tex"
	VPNTTB   = "pos.normal.tex.tangent.bitangent"
	VPNTT    = "pos.normal.tex.tex2"
	VPNTWJ   = "pos.normal.tex.weight.joint"
	VPNTTBWJ = "pos.normal.tex.tangent.bitangent.weight.joint"
)

type VertexCacheManager struct {
	ActiveSkin    int
	RenderManager *RenderManager
	CacheId       int
	Strides       map[string]int32
	Formats       map[string][]VertexFormat
	Caches        map[string][MAX_CACHES]*VertexCache
	StaticBuffers []*StaticBuffer
}

func NewVertexCacheManager(
	renderManager *RenderManager,
	maxVerts, maxIndices, maxDrawCommands, maxInstances int32,
) *VertexCacheManager {
	vm := &VertexCacheManager{
		ActiveSkin:    -1,
		RenderManager: renderManager,
		Caches:        make(map[string][MAX_CACHES]*VertexCache),
		Formats:       make(map[string][]VertexFormat),
		StaticBuffers: make([]*StaticBuffer, 0),
		Strides:       make(map[string]int32),
	}
	vm.Strides[VP] = 12
	vm.Strides[VPNT] = 32
	vm.Strides[VPNTT] = 40
	vm.Strides[VPNTWJ] = 56
	vm.Strides[VPNTTB] = 56
	vm.Strides[VPNTTBWJ] = 80

	vm.Formats[VP] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: 0,
		},
	}
	vm.Formats[VPNT] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: 0,
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 6),
		},
	}
	vm.Formats[VPNTT] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: 0,
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 6),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 8),
		},
	}
	vm.Formats[VPNTWJ] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: 0,
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 6),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(uint32(0)) * 8),
		},
		{
			ComponentSize:  3,
			Type:           gl.UNSIGNED_INT,
			RelativeOffset: uint32(unsafe.Sizeof(uint32(0)) * 11),
		},
	}
	vm.Formats[VPNTTB] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: 0,
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 6),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 8),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 11),
		},
	}
	vm.Formats[VPNTTBWJ] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: 0,
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 6),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 8),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 11),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 14),
		},
		{
			ComponentSize:  3,
			Type:           gl.UNSIGNED_INT,
			RelativeOffset: uint32(unsafe.Sizeof(uint32(0)) * 17),
		},
	}
	vm.Caches[VP] = [MAX_CACHES]*VertexCache{}
	vm.Caches[VPNT] = [MAX_CACHES]*VertexCache{}
	vm.Caches[VPNTT] = [MAX_CACHES]*VertexCache{}
	vm.Caches[VPNTTB] = [MAX_CACHES]*VertexCache{}
	vm.Caches[VPNTWJ] = [MAX_CACHES]*VertexCache{}
	vm.Caches[VPNTTBWJ] = [MAX_CACHES]*VertexCache{}
	for i := range MAX_CACHES {
		c := vm.Caches[VP]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			vm,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxInstances,
			vm.Strides[VP],
			-1,
			vm.CacheId,
			vm.Formats[VP],
		)
		vm.Caches[VP] = c
		vm.CacheId++
		c = vm.Caches[VPNTT]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			vm,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxInstances,
			vm.Strides[VPNTT],
			-1,
			vm.CacheId,
			vm.Formats[VPNTT],
		)
		vm.Caches[VPNTT] = c
		vm.CacheId++
		c = vm.Caches[VPNT]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			vm,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxInstances,
			vm.Strides[VPNT],
			-1,
			vm.CacheId,
			vm.Formats[VPNT],
		)
		vm.Caches[VPNT] = c
		vm.CacheId++
		c = vm.Caches[VPNTWJ]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			vm,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxInstances,
			vm.Strides[VPNTWJ],
			-1,
			vm.CacheId,
			vm.Formats[VPNTWJ],
		)
		vm.Caches[VPNTWJ] = c
		vm.CacheId++
		c = vm.Caches[VPNTTBWJ]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			vm,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxInstances,
			vm.Strides[VPNTTBWJ],
			-1,
			vm.CacheId,
			vm.Formats[VPNTTBWJ],
		)
		vm.Caches[VPNTTBWJ] = c
		vm.CacheId++
		c = vm.Caches[VPNTTB]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			vm,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxInstances,
			vm.Strides[VPNTTB],
			-1,
			vm.CacheId,
			vm.Formats[VPNTTB],
		)
		vm.Caches[VPNTTB] = c
	}
	return vm
}

func (vm *VertexCacheManager) Destory() {
	for _, cs := range vm.Caches {
		for _, c := range cs {
			c.Destroy()
		}
	}
	for _, ss := range vm.StaticBuffers {
		ss.Destory()
	}
	clear(vm.StaticBuffers)
	clear(vm.Caches)
}

func (vm *VertexCacheManager) LoadMatrix(vertexType string,
	baseInstance int,
	skinId int, world []y3d.Mat4) error {
	vc, ok := vm.Caches[vertexType]
	if !ok {
		return errors.New("invalid vertex type")
	}
	for i := range MAX_CACHES {
		if vc[i].SkinId == skinId {
			vc[i].LoadMatrix(
				baseInstance, world)
			return nil
		}
	}
	return nil
}

func (vm *VertexCacheManager) LoadCache(
	vertexType string,
	dataV, dataI *bytes.Buffer,
	skinID int,
	command *DrawCommand,
) error {
	var empty, fullest *VertexCache
	vc, ok := vm.Caches[vertexType]
	if !ok {
		return errors.New("invalid vertex type")
	}
	fullest = vc[0]
	for i := range MAX_CACHES {
		if vc[i].SkinId == skinID {
			return vc[i].Add(
				command,
				int(command.InstanceCount),
				dataV,
				dataI,
			)
		}
		if vc[i].IsEmpty() {
			empty = vc[i]
		}
		if vc[i].NumVertics > fullest.NumVertics {
			fullest = vc[i]
		}
	}
	if empty != nil {
		empty.SetSkin(skinID)
		return empty.Add(
			command,
			int(command.InstanceCount),
			dataV,
			dataI,
		)
	}
	fullest.SetSkin(skinID)
	return fullest.Add(
		command,
		int(command.InstanceCount),
		dataV,
		dataI,
	)
}

func (vm *VertexCacheManager) ForceRender(vertexType string) error {
	vc, ok := vm.Caches[vertexType]
	log.Println(vertexType)
	if !ok {
		return errors.New("invalid vertex type")
	}
	for _, c := range vc {
		c.Render()
	}
	return nil
}

func (vm *VertexCacheManager) ForceRenderAll() error {
	for _, vc := range vm.Caches {
		for _, c := range vc {
			c.Render()
		}
	}
	return nil
}

func (vm *VertexCacheManager) CreateStaticBuffer(
	vertexType string,
	dataV, dataI *bytes.Buffer,
	command []DrawCommand,
	skinId int,
	instanceCount int,
	idxType uint32,
) (int, error) {
	id := len(vm.StaticBuffers)
	s := NewStaticBuffer(
		vm,
		dataV, dataI,
		command,
		instanceCount,
		skinId,
		vm.Strides[vertexType],
		vm.Formats[vertexType],
		idxType,
		id,
	)
	if s == nil {
		return -1, errors.New("cannot create static buffer")
	}
	vm.StaticBuffers = append(vm.StaticBuffers, s)
	return id, nil
}

func (vm *VertexCacheManager) RenderSB(id int, world []y3d.Mat4) {
	if id >= len(vm.StaticBuffers) {
		panic("invalid static buffer id")
	}
	vm.StaticBuffers[id].Render(world)
}

func (vm *VertexCacheManager) CreateDrawCommand(dataI *bytes.Buffer,
	instanceCount int) DrawCommand {
	return DrawCommand{
		VertexCount:   uint32(dataI.Len() / 4),
		InstanceCount: uint32(instanceCount),
		FirstIndex:    0,
		BaseVertex:    0,
		BaseInstance:  0,
	}
}

// vertices are shared between triangles, this bitangent/tangent calculation would not work
// have to render the data into an agency list or a new data structure, then build the tangent
// then re-render it into a vertex list with the indices into that list
func (vm *VertexCacheManager) CalculateTangentSpace(dataV, dataI *bytes.Buffer, vertexType string) (*bytes.Buffer, error) {
	if vertexType != VPNTTB && vertexType != VPNTTBWJ {
		return nil, fmt.Errorf("cannot create tangent space for this format: %s", vertexType)
	}
	format := vm.Formats[vertexType]
	stride := vm.Strides[vertexType]
	b := dataV.Bytes()
	getPos := func(i int) y3d.Vec3 {
		off := (int(stride) * i) + int(format[0].RelativeOffset)
		end := off + int(format[0].ComponentSize)*int(unsafe.Sizeof(float32(0)))
		v := b[off:end]
		vp := unsafe.Slice((*float32)(unsafe.Pointer(unsafe.SliceData(v))), 3)
		return y3d.Vec3{
			X: vp[0],
			Y: vp[1],
			Z: vp[2],
		}
	}

	getUV := func(i int) y3d.Vec2 {
		off := (int(stride) * i) + int(format[2].RelativeOffset)
		end := off + int(format[2].ComponentSize)*int(unsafe.Sizeof(float32(0)))
		v := b[off:end]
		vp := unsafe.Slice((*float32)(unsafe.Pointer(unsafe.SliceData(v))), 2)
		return y3d.Vec2{
			X: vp[0],
			Y: vp[1],
		}
	}
	writeTB := func(T, B []float32, i int) {
		off := (int(stride) * i) + int(format[0].RelativeOffset)
		end := off + int(format[0].ComponentSize)*int(unsafe.Sizeof(float32(0)))
		vt := b[off:end]
		t := unsafe.Slice((*float32)(unsafe.Pointer(unsafe.SliceData(vt))), 3)

		vb := b[off:end]
		b := unsafe.Slice((*float32)(unsafe.Pointer(unsafe.SliceData(vb))), 3)

		copy(t, T)
		copy(b, B)
	}
	count := dataI.Len() / 4
	i := dataI.Bytes()
	indices := unsafe.Slice((*uint32)(unsafe.Pointer(&i[0])), count)
	for i := 0; i < count-3; i += 3 {
		var T, B [3]y3d.Vec3
		i1, i2, i3 := indices[i], indices[i+1], indices[i+2]
		pos1 := getPos(int(i1))
		pos2 := getPos(int(i2))
		pos3 := getPos(int(i3))

		uv1 := getUV(int(i1))
		uv2 := getUV(int(i2))
		uv3 := getUV(int(i3))

		f1 := y3d.Vec2{X: uv2.X - uv1.X, Y: uv2.Y - uv1.Y}
		f2 := y3d.Vec2{X: uv3.X - uv1.X, Y: uv3.Y - uv1.Y}

		vcA := y3d.Vec3{
			X: pos2.X - pos1.X,
			Y: f1.X,
			Z: f1.Y,
		}
		vcB := y3d.Vec3{
			X: pos3.X - pos1.X,
			Y: f2.X,
			Z: f2.Y,
		}
		vc := y3d.Cross(vcA, vcB)
		if vc.X != 0.0 {
			T[0].X = -vc.Y / vc.X
			B[0].X = -vc.Z / vc.X
		}

		vcA = y3d.Vec3{
			X: pos2.Y - pos1.Y,
			Y: f1.X,
			Z: f1.Y,
		}
		vcB = y3d.Vec3{
			X: pos3.Y - pos1.Y,
			Y: f2.X,
			Z: f2.Y,
		}
		vc = y3d.Cross(vcA, vcB)
		if vc.X != 0.0 {
			T[0].Y = -vc.Y / vc.X
			B[0].Y = -vc.Z / vc.X
		}

		vcA = y3d.Vec3{
			X: pos2.Z - pos1.Z,
			Y: f1.X,
			Z: f1.Y,
		}
		vcB = y3d.Vec3{
			X: pos3.Z - pos1.Z,
			Y: f2.X,
			Z: f2.Y,
		}
		vc = y3d.Cross(vcA, vcB)
		if vc.X != 0.0 {
			T[0].Z = -vc.Y / vc.X
			B[0].Z = -vc.Z / vc.X
		}
		T[0] = y3d.Normalize(T[0])
		B[0] = y3d.Normalize(B[0])
		T[1] = T[0]
		T[2] = T[0]
		B[1] = B[0]
		B[2] = B[0]

		writeTB(unsafe.Slice((*float32)(unsafe.Pointer(&T[0])), 3),
			unsafe.Slice((*float32)(unsafe.Pointer(&B[0])), 3), int(i1))
		writeTB(unsafe.Slice((*float32)(unsafe.Pointer(&T[1])), 3),
			unsafe.Slice((*float32)(unsafe.Pointer(&B[1])), 3), int(i2))
		writeTB(unsafe.Slice((*float32)(unsafe.Pointer(&T[2])), 3),
			unsafe.Slice((*float32)(unsafe.Pointer(&B[2])), 3), int(i3))
	}
	return dataV, nil
}

func (vm *VertexCacheManager) GetScalingAndBox(dataV, dataI *bytes.Buffer, scale float32, vertexType string) (float32, y3d.AABB) {
	if scale == 0.0 {
		return 0.0, y3d.UnitAABB
	}
	format := vm.Formats[vertexType]
	stride := vm.Strides[vertexType]
	b := dataV.Bytes()
	getPos := func(i int) y3d.Vec3 {
		off := (int(stride) * i) + int(format[0].RelativeOffset)
		end := off + int(format[0].ComponentSize)*int(unsafe.Sizeof(float32(0)))
		v := b[off:end]
		vp := unsafe.Slice((*float32)(unsafe.Pointer(unsafe.SliceData(v))),
			format[0].ComponentSize)
		return y3d.Vec3{
			X: vp[0],
			Y: vp[1],
			Z: vp[2],
		}
	}
	box := y3d.AABB{
		Min: y3d.Vec3{
			X: 999999.99,
			Y: 999999.99,
			Z: 999999.99,
		},
		Max: y3d.Vec3{
			X: -999999.99,
			Y: -999999.99,
			Z: -999999.99,
		},
	}
	count := dataI.Len() / 4
	i := dataI.Bytes()
	indices := unsafe.Slice((*uint32)(unsafe.Pointer(unsafe.SliceData(i))), count)
	for _, i := range indices {
		vertex := getPos(int(i))
		box.Max.X = max(vertex.X, box.Max.X)
		box.Max.Y = max(vertex.Y, box.Max.Y)
		box.Max.Z = max(vertex.Z, box.Max.Z)

		box.Min.X = min(vertex.X, box.Min.X)
		box.Min.Y = min(vertex.Y, box.Min.Y)
		box.Min.Z = min(vertex.Z, box.Min.Z)
	}
	scaling := (box.Max.Y - box.Min.Y) / scale
	return 1.0 / scaling, box
}
