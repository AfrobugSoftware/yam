package ygame

import (
	"bytes"
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

type InputFunc func(*ycore.InputManager, ycontroller.Controller)

var (
	MoveControllerComponent = ycore.ComponentId(1)
	DrawCommandComponent    = ycore.ComponentId(2)
	TransformComponent      = ycore.ComponentId(3)
	SkinComponent           = ycore.ComponentId(4)
	VertexTypeComponent     = ycore.ComponentId(5)
	RenderFlagComponent     = ycore.ComponentId(6)
	BoundingBoxComponent    = ycore.ComponentId(7)
	EntityTagComponent      = ycore.ComponentId(9)
	InstanceComponent       = ycore.ComponentId(10)
	InputComponent          = ycore.ComponentId(11)
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

func (l *Level) RenderEntities(r *ycore.RenderManager, cull bool) {
	vm := r.VertextManager
	frustum := r.GetFrustum()
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
		c := box.Cull(frustum[:])
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
		v.LoadMatrix(int(drawcommand.BaseInstance), transform.World) //this needs to be world
		v.AddRenderFlag(int(drawcommand.BaseInstance), rflag)
	}
}

func (l *Level) UpdateEntitesMovement(deltaTime float32) {
	entites := l.World.Query(TransformComponent, BoundingBoxComponent, MoveControllerComponent)
	for _, e := range entites {
		movement := l.World.GetComponent(e, MoveControllerComponent).(*ycontroller.MovementController)
		transform := l.World.GetComponent(e, TransformComponent).(ycore.Transform)
		box := l.World.GetComponent(e, BoundingBoxComponent).(y3d.AABB)
		movement.Update(deltaTime)
		transform.SetPositionOrientation(movement.Position, movement.Orientation)
		transform.World = transform.Local

		box = transform.TransformAABB(box)
		l.World.SetComponent(e, TransformComponent, transform)
		l.World.SetComponent(e, BoundingBoxComponent, box)
	}
}

func (l *Level) ProcessEntityInput(input *ycore.InputManager) {
	entites := l.World.Query(InputComponent, MoveControllerComponent)
	for _, e := range entites {
		movement := l.World.GetComponent(e, MoveControllerComponent).(*ycontroller.MovementController)
		inputfunc := l.World.GetComponent(e, InputComponent).(InputFunc)

		inputfunc(input, movement)
	}
}

func (l *Level) AddEntityDrawData(
	ent ycore.EntityId,
	entityTag uint32,
	vertexType string,
	skinId int,
	box y3d.AABB,
	v, i *bytes.Buffer, vm *ycore.VertexCacheManager) error {
	cmd := ycore.DrawCommand{}
	err := vm.LoadCache(vertexType, v, i, skinId, &cmd)
	if err != nil {
		return err
	}
	l.World.AddComponent(ent, EntityTagComponent, entityTag)
	l.World.AddComponent(ent, DrawCommandComponent, cmd)
	l.World.AddComponent(ent, SkinComponent, skinId)
	l.World.AddComponent(ent, RenderFlagComponent, uint32(0))
	l.World.AddComponent(ent, InstanceComponent, 1)
	l.World.AddComponent(ent, VertexTypeComponent, vertexType)
	l.World.AddComponent(ent, BoundingBoxComponent, box)
	return nil
}

func (l *Level) AddEntityMoveData(
	ent ycore.EntityId,
	pos y3d.Vec3,
	orientation y3d.Quaternion,
) {
	move := ycontroller.NewMovementController(pos, orientation)
	transform := *ycore.NewTransform()
	l.World.AddComponent(ent, TransformComponent, transform)
	l.World.AddComponent(ent, MoveControllerComponent, move)
}

func (l *Level) AddEntityTransformData(ent ycore.EntityId,
	transform ycore.Transform) {
	l.World.AddComponent(ent, TransformComponent, transform)
}

func (l *Level) AddEntityInputData(ent ycore.EntityId,
	input InputFunc) {
	l.World.AddComponent(ent, InputComponent, input)
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
