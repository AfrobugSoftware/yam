package ycore

import (
	"bytes"
	"yam/y3d"
	"yam/ycontroller"
)

type InputFunc func(*InputManager, ycontroller.Controller)

const (
	MoveControllerComponent ComponentId = iota
	DrawCommandComponent
	TransformComponent
	SkinComponent
	VertexTypeComponent
	RenderStateComponent
	BoundingBoxComponent
	EntityTagComponent
	InputComponent
)

func AddEntityDrawData(
	world *World,
	ent EntityId,
	entityTag uint32,
	vertexType string,
	skinId int,
	box y3d.AABB,
	v, i *bytes.Buffer, vm *VertexCacheManager) error {
	cmd := DrawCommand{}
	err := vm.LoadCache(vertexType, v, i, skinId, &cmd)
	if err != nil {
		return err
	}
	world.AddComponent(ent, EntityTagComponent, entityTag)
	world.AddComponent(ent, DrawCommandComponent, cmd)
	world.AddComponent(ent, SkinComponent, skinId)
	world.AddComponent(ent, RenderStateComponent, PerObjectRendeState{
		Alpha: 1.0,
	})
	world.AddComponent(ent, VertexTypeComponent, vertexType)
	world.AddComponent(ent, BoundingBoxComponent, box)
	return nil
}

func AddEntityMoveData(
	world *World,
	ent EntityId,
	pos y3d.Vec3,
	orientation y3d.Quaternion,
) {
	move := ycontroller.NewMovementController(pos, orientation)
	transform := *NewTransform()
	world.AddComponent(ent, TransformComponent, transform)
	world.AddComponent(ent, MoveControllerComponent, move)
}

func AddEntityTransformData(
	world *World,
	ent EntityId,
	transform Transform) {
	world.AddComponent(ent, TransformComponent, transform)
}

func AddEntityInputData(
	world *World,
	ent EntityId,
	input InputFunc) {
	world.AddComponent(ent, InputComponent, input)
}
