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

// make the instance an entity, so entites can be instances
func AddEntityMatInstnaces(world *World, e EntityId, mats ...y3d.Mat4) {
	imat := make([]y3d.Mat4, 0, len(mats))
	imat = append(imat, mats...)

	world.AddComponent(e, MatInstanceComponent, imat)
}

func AddEnityRenderParamsInstance(world *World, e EntityId, p ...PerObjectRenderState) {
	ip := make([]PerObjectRenderState, 0, len(p))
	ip = append(ip, p...)
	world.AddComponent(e, RenderParamsInstanceComponent, ip)
}
