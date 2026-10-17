package ygame

import (
	"encoding/gob"
	"errors"
	"os"
	"yam/y3d"
	"yam/ycontroller"
	"yam/ycore"
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
	RenderFlagComponent     = ycore.RegisterComponent[uint32]()
	BoundingBoxComponent    = ycore.RegisterComponent[y3d.AABB]()
	EntityTypeComponent     = ycore.RegisterComponent[uint32]()
	InstanceComponent       = ycore.RegisterComponent[int]()
)

type LevelParameters struct {
	DeltaTime float32
	GameTime  float32
}

type Level struct {
	World   *ycore.World
	OctTree *y3d.Octree
	Meshes  []*Polymesh

	levelParametersUBO uint32
}

func NewLevel() *Level {
	return &Level{
		World: ycore.NewWorld(),
	}
}

func (l *Level) RenderEntities(vm *ycore.VertexCacheManager, frustum []y3d.Plane, cull bool) {
	entities := l.World.Query(
		DrawCommandComponent,
		RenderFlagComponent,
		TransformComponent,
		SkinComponent,
		InstanceComponent,
		VertexTypeComponent,
		BoundingBoxComponent)
	for _, e := range entities {
		box := l.World.GetComponent(e, BoundingBoxComponent).(y3d.AABB)
		c := box.Cull(frustum)
		if cull && c == y3d.CULLED {
			continue
		}
		vertexType := l.World.GetComponent(e, VertexTypeComponent).(string)
		skinId := l.World.GetComponent(e, SkinComponent).(int)
		v, err := vm.GetCache(vertexType, skinId)
		if err != nil {
			continue //no cache for this entity
		}
		instance := l.World.GetComponent(e, InstanceComponent).(int)
		transform := l.World.GetComponent(e, TransformComponent).(ycore.Transform)
		rflag := l.World.GetComponent(e, RenderFlagComponent).(uint32)
		drawcommand := l.World.GetComponent(e, DrawCommandComponent).(ycore.DrawCommand)

		v.AddInstances(&drawcommand, instance)
		v.AddDrawCommand(drawcommand)
		v.LoadMatrix(int(drawcommand.BaseInstance), transform.World)
		v.AddRenderFlag(int(drawcommand.BaseInstance), rflag)
	}
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
