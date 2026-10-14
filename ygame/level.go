package ygame

import (
	"encoding/gob"
	"errors"
	"os"
	"unsafe"
	"yam/y3d"
	"yam/ycontroller"
	"yam/ycore"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	LOB_NONE          = 0
	LOB_ALL           = 2
	LOB_SECTOR        = 4
	LOB_MESH          = 8
	LOB_POLYGON       = 16
	LOB_PORTAL        = 32
	LOB_ENTITY        = 64
	LOB_LIGH          = 128
	LOB_SPAWNPOINT    = 256
	LOB_TRIGGER_POINT = 512
)

const (
	MAX_LIGHTS = 10
)

type AXIS int

const (
	X_AXIS AXIS = iota
	Y_AXIS
	Z_AXIS
)

var (
	MoveControllerComponent = ycore.RegisterComponent[ycontroller.MovementController]()
	DrawCommandComponent    = ycore.RegisterComponent[ycore.DrawCommand]()
	TransformComponent      = ycore.RegisterComponent[ycore.Transform]()
	SkinComponent           = ycore.RegisterComponent[int]()
	VertexTypeComponent     = ycore.RegisterComponent[string]()
	BoundingBoxComponent    = ycore.RegisterComponent[y3d.AABB]()
	EntityTypeComponent     = ycore.RegisterComponent[uint32]()
)

type LevelParameters struct {
}

type Level struct {
	World   *ycore.World
	OctTree *y3d.Octree

	lightSSBO          uint32
	levelParametersUBO uint32
	numLight           int
}

func NewLevel() *Level {
	var ssbo uint32
	gl.CreateBuffers(1, &ssbo)
	gl.NamedBufferStorage(ssbo, int(unsafe.Sizeof(y3d.Light{})*MAX_LIGHTS), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	return &Level{
		World:     ycore.NewWorld(),
		lightSSBO: ssbo,
	}
}
func (l *Level) AddLight(light y3d.Light) {
	if l.numLight >= MAX_LIGHTS {
		return
	}
	ptr := unsafe.Pointer(&light)
	size := unsafe.Sizeof(light)
	s := unsafe.Slice((*byte)(ptr), int(size))
	gl.NamedBufferSubData(l.lightSSBO, (int(size) * l.numLight), int(size), gl.Ptr(s))
}
func (l *Level) UseLight() {
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, ycore.LIGHT_BINDING, l.lightSSBO)
}
func (l *Level) ClearLight() {
	l.numLight = 0
}
func (l *Level) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	if encoder == nil {
		return errors.New("cannot create encoder")
	}
	l.World.Save(encoder)
	return nil
}

func (l *Level) Load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(file)
	if decoder == nil {
		return errors.New("cannot create decoder")
	}
	l.World.Load(decoder)
	return nil
}

func (l *Level) Destory() {
	gl.DeleteBuffers(1, &l.lightSSBO)
}
