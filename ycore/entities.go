package ycore

import (
	"bytes"
	"yam/y3d"
	"yam/ycontroller"
)

type InputFunc func(*InputManager, ycontroller.Controller)

const (
	MoveControllerComponent ComponentId = iota
	TransformComponent
	BoundingBoxComponent
	EntityTagComponent
	InputComponent
	RenderableComponent
	SkeletonComponent
	MatInstanceComponent
	RenderParamsInstanceComponent
	HierarchyComponent
	FlagComponent
	AnimationComponent
	InstanceComponent
)

func AddEntityRenderableData(
	world *World,
	ent EntityId,
	entityTag uint32,
	vertexType string,
	skinId int,
	v, i *bytes.Buffer, vm *VertexCacheManager) error {
	cmd := DrawCommand{}
	err := vm.LoadCache(vertexType, v, i, skinId, &cmd)
	if err != nil {
		return err
	}
	renderable := Renderable{
		DrawCmd:    cmd,
		VertexType: vertexType,
		SkinId:     skinId,
		RenderParams: PerObjectRenderState{
			Alpha: 1.0,
		},
	}
	box := NewBox()
	_, localBox := vm.GetScalingAndBox(v, i, 1.0, vertexType)
	box.Local = localBox
	world.AddComponent(ent, EntityTagComponent, entityTag)
	world.AddComponent(ent, BoundingBoxComponent, box)
	world.AddComponent(ent, RenderableComponent, renderable)
	return nil
}

func AddEntityMoveData(
	world *World,
	ent EntityId,
	pos y3d.Vec3,
	orientation y3d.Quaternion,
) *ycontroller.MovementController {
	move := ycontroller.NewMovementController(pos, orientation)
	world.AddComponent(ent, TransformComponent, NewTransform())
	world.AddComponent(ent, MoveControllerComponent, move)
	return move
}

func AddEntityAsInstance(w *World, holder EntityId, instance EntityId) {
	if !w.HasComponent(holder, InstanceComponent) {
		instances := make([]EntityId, 0)
		instances = append(instances, instance)
		box := w.GetComponent(holder, BoundingBoxComponent).(Box)
		w.AddComponent(instance, BoundingBoxComponent, box)
		w.AddComponent(holder, InstanceComponent, instances)
		return
	}
	instances := w.GetComponent(holder, InstanceComponent).([]EntityId)
	instances = append(instances, instance)
	box := w.GetComponent(holder, BoundingBoxComponent).(Box)
	w.AddComponent(instance, BoundingBoxComponent, box)
	w.SetComponent(holder, InstanceComponent, instances)
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
