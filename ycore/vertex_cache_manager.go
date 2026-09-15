package ycore

import (
	"bytes"
	"errors"
	"unsafe"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	INVALID_CACHE = -1
	MAX_CACHES    = 10
)

const (
	VP     = "pos"
	VPNT   = "pos.normal.tex"
	VPNTT  = "pos.normal.tex.tex2"
	VPNTWJ = "pos.normal.tex.joint.weight"
)

type VertexCacheManager struct {
	ActiveCache   int
	ActiveSB      int
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
	maxVerts, maxIndices, maxDrawCommands, maxWorldMatrix int32,
) *VertexCacheManager {
	vm := &VertexCacheManager{
		ActiveCache:   INVALID_CACHE,
		ActiveSkin:    -1,
		ActiveSB:      -1,
		RenderManager: renderManager,
		Caches:        make(map[string][MAX_CACHES]*VertexCache),
		Formats:       make(map[string][]VertexFormat),
		StaticBuffers: make([]*StaticBuffer, 0),
		Strides:       make(map[string]int32),
	}
	vm.Strides[VP] = int32(12)
	vm.Strides[VPNT] = int32(32)
	vm.Strides[VPNTT] = int32(40)
	vm.Strides[VPNTWJ] = int32(56)

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
	for i := range MAX_CACHES {
		c := vm.Caches[VP]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxWorldMatrix,
			vm.Strides[VP],
			-1,
			vm.CacheId,
			vm.Formats[VP],
		)
		vm.CacheId++
		c = vm.Caches[VPNTT]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxWorldMatrix,
			vm.Strides[VPNTT],
			-1,
			vm.CacheId,
			vm.Formats[VPNTT],
		)
		vm.CacheId++
		c = vm.Caches[VPNT]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxWorldMatrix,
			vm.Strides[VPNT],
			-1,
			vm.CacheId,
			vm.Formats[VPNT],
		)
		vm.CacheId++
		c = vm.Caches[VPNTWJ]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			maxVerts,
			maxIndices,
			maxDrawCommands,
			maxWorldMatrix,
			vm.Strides[VPNTWJ],
			-1,
			vm.CacheId,
			vm.Formats[VPNTWJ],
		)
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

func (vm *VertexCacheManager) Render(
	vertexType string,
	dataV, dataI *bytes.Buffer,
	skinID int,
	world []y3d.Mat4,
	command DrawCommand,
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
				world,
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
			world,
			dataV,
			dataI,
		)
	}
	fullest.SetSkin(skinID)
	return fullest.Add(
		command,
		world,
		dataV,
		dataI,
	)
}

func (vm *VertexCacheManager) ForceFlush(vertexType string) error {
	vc, ok := vm.Caches[vertexType]
	if !ok {
		return errors.New("invalid vertex type")
	}
	for _, c := range vc {
		c.Flush()
	}
	return nil
}

func (vm *VertexCacheManager) ForceFlushAll() error {
	for _, vc := range vm.Caches {
		for _, c := range vc {
			c.Flush()
		}
	}
	return nil
}

func (vm *VertexCacheManager) CreateStaticBuffer(
	vertexType string,
	dataV, dataI *bytes.Buffer,
	command []DrawCommand,
	skinId int,
	world []y3d.Mat4,
) (int, error) {
	id := len(vm.StaticBuffers)
	s := NewStaticBuffer(
		vm,
		dataV, dataI,
		command,
		skinId,
		world,
		vm.Strides[vertexType],
		vm.Formats[vertexType],
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
