package ycore

import (
	"bytes"
	"yam/y3d"
	"yam/ycontroller"
)

type InputFunc func(*InputManager, ycontroller.Controller)

const (
	MoveControllerComponent = ComponentId(1)
	DrawCommandComponent    = ComponentId(2)
	TransformComponent      = ComponentId(3)
	SkinComponent           = ComponentId(4)
	VertexTypeComponent     = ComponentId(5)
	RenderFlagComponent     = ComponentId(6)
	BoundingBoxComponent    = ComponentId(7)
	EntityTagComponent      = ComponentId(9)
	InstanceComponent       = ComponentId(10)
	InputComponent          = ComponentId(11)
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
	world.AddComponent(ent, RenderFlagComponent, uint32(0))
	world.AddComponent(ent, InstanceComponent, 1)
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
