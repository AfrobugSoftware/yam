package ycore

import (
	"bytes"
	"yam/y3d"
	"yam/ycontroller"
	"yam/yphy"
)

type InputFunc func(*InputManager, ycontroller.Controller)

const (
	MoveControllerComponent ComponentId = iota
	TransformComponent
	BoundingBoxComponent
	InputComponent
	RenderableComponent
	SkeletonComponent
	HierarchyComponent
	FlagComponent
	AnimationComponent
	InstanceComponent
	ScaleComponent
	PbrMaterialComponent
)

func AddEntityRenderableData(
	world *World,
	ent EntityId,
	vertexType string,
	scale float32,
	v, i *bytes.Buffer, vm *VertexCacheManager) error {
	cmd := DrawCommand{}
	vid, err := vm.LoadCache(vertexType, v, i, &cmd)
	if err != nil {
		return err
	}
	renderable := Renderable{
		DrawCmd:       cmd,
		VertexType:    vertexType,
		VertexCacheId: vid,
		RenderParams: PerObjectRenderState{
			Alpha: 1.0,
		},
	}
	format := vm.Formats[vertexType]
	stride := vm.Strides[vertexType]
	box := yphy.BoundBox(int(stride), int(format[0].RelativeOffset), v, i)
	world.AddComponent(ent, BoundingBoxComponent, box)
	world.AddComponent(ent, RenderableComponent, renderable)
	world.AddComponent(ent, ScaleComponent, scale)
	return nil
}

func AddScaleCompoment(world *World,
	ent EntityId, scale float32) {
	world.AddComponent(ent, ScaleComponent, scale)
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
		box := w.GetComponent(holder, BoundingBoxComponent).(yphy.Box)
		w.AddComponent(instance, BoundingBoxComponent, box)
		w.AddComponent(holder, InstanceComponent, instances)
		return
	}
	instances := w.GetComponent(holder, InstanceComponent).([]EntityId)
	instances = append(instances, instance)
	box := w.GetComponent(holder, BoundingBoxComponent).(yphy.Box)
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
