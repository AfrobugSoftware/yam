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

func (l *Level) UpdateEntitesMovement(deltaTime float32) {
	entites := l.World.Query(
		ycore.TransformComponent,
		ycore.BoundingBoxComponent,
		ycore.MoveControllerComponent)
	for _, e := range entites {
		movement := l.World.GetComponent(e, ycore.MoveControllerComponent).(*ycontroller.MovementController)
		transform := l.World.GetComponent(e, ycore.TransformComponent).(ycore.Transform)
		box := l.World.GetComponent(e, ycore.BoundingBoxComponent).(y3d.AABB)
		movement.Update(deltaTime)
		transform.SetPositionOrientation(movement.Position, movement.Orientation)
		transform.World = transform.Local

		box = transform.TransformAABB(box)
		l.World.SetComponent(e, ycore.TransformComponent, transform)
		l.World.SetComponent(e, ycore.BoundingBoxComponent, box)
	}
}

func (l *Level) ProcessEntityInput(input *ycore.InputManager) {
	entites := l.World.Query(ycore.InputComponent, ycore.MoveControllerComponent)
	for _, e := range entites {
		movement := l.World.GetComponent(e, ycore.MoveControllerComponent).(*ycontroller.MovementController)
		inputfunc := l.World.GetComponent(e, ycore.InputComponent).(ycore.InputFunc)

		inputfunc(input, movement)
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
