package ymanager

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
	VPNT   = "pos.normal.tex"
	VPNTT  = "pos.normal.tex.tex2"
	VPNTWJ = "pos.normal.tex.joint.weight"
)

type VertexCacheManager struct {
	ActiveCache   int
	ActiveSkin    int
	RenderManager *RenderManager
	CacheId       int
	Strides       map[string]int32
	Formats       map[string][]VertexFormat
	Caches        map[string][MAX_CACHES]*VertexCache
	StaticBuffers map[string][]*StaticBuffer
}

func NewVertexCacheManager(
	renderManager *RenderManager,
	maxVerts, maxIndices, maxDrawCommands int32,
) *VertexCacheManager {
	vm := &VertexCacheManager{
		ActiveCache:   INVALID_CACHE,
		ActiveSkin:    -1,
		RenderManager: renderManager,
		Caches:        make(map[string][MAX_CACHES]*VertexCache),
		Formats:       make(map[string][]VertexFormat),
		StaticBuffers: make(map[string][]*StaticBuffer),
		Strides:       make(map[string]int32),
	}
	vm.Strides[VPNT] = int32(32)
	vm.Strides[VPNTT] = int32(40)
	vm.Strides[VPNTWJ] = int32(56)
	vm.Formats[VPNT] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 2),
		},
	}
	vm.Formats[VPNTT] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 2),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 2),
		},
	}
	vm.Formats[VPNTWJ] = []VertexFormat{
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  2,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 2),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
		{
			ComponentSize:  3,
			Type:           gl.FLOAT,
			RelativeOffset: uint32(unsafe.Sizeof(float32(0)) * 3),
		},
	}
	for i := range MAX_CACHES {
		c := vm.Caches[VPNTT]
		c[i] = NewVertexCache(
			renderManager.SkinManager,
			maxVerts,
			maxIndices,
			maxDrawCommands,
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
			vm.Strides[VPNTWJ],
			-1,
			vm.CacheId,
			vm.Formats[VPNTWJ],
		)
	}
	vm.StaticBuffers[VPNT] = make([]*StaticBuffer, 0)
	vm.StaticBuffers[VPNTT] = make([]*StaticBuffer, 0)
	vm.StaticBuffers[VPNTWJ] = make([]*StaticBuffer, 0)
	return vm
}

func (vm *VertexCacheManager) Destory() {
	for _, cs := range vm.Caches {
		for _, c := range cs {
			c.Destroy()
		}
	}
	for _, ss := range vm.StaticBuffers {
		for _, s := range ss {
			s.Destory()
		}
	}
	clear(vm.StaticBuffers)
	clear(vm.Caches)
}

func (vm *VertexCacheManager) Render(
	vertexType string,
	dataV, dataI *bytes.Buffer,
	skinID int,
	world y3d.Mat4,
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
	s, ok := vm.StaticBuffers[vertexType]
	if !ok {
		return -1, errors.New("invalid vertex type")
	}
	s = append(s, NewStaticBuffer(
		vm,
		dataV, dataI,
		command,
		skinId,
		world,
		vm.Strides[vertexType],
		vm.Formats[vertexType],
	))
	vm.StaticBuffers[vertexType] = s
	return len(s) - 1, nil
}
