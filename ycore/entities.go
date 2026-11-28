package ycore

import (
	"bytes"
	"slices"
	"yam/y3d"
	"yam/yphy"
	"yam/ywd"
)

func AddEntityRenderableData(
	world *ywd.World,
	ent ywd.EntityId,
	vertexType string,
	scale y3d.Vec3,
	v, i, p *bytes.Buffer, vm *VertexCacheManager) error {
	cmd := DrawCommand{}
	vid, err := vm.LoadCache(vertexType, v, i, p, &cmd)
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
	world.AddComponent(ent, ywd.BoundingBoxComponent, box)
	world.AddComponent(ent, ywd.RenderableComponent, renderable)
	world.AddComponent(ent, ywd.ScaleComponent, scale)
	return nil
}

func AddScaleCompoment(world *ywd.World,
	ent ywd.EntityId, scale float32) {
	world.AddComponent(ent, ywd.ScaleComponent, y3d.Vec3{X: scale, Y: scale, Z: scale})
}

func AddEntityMoveData(
	world *ywd.World,
	ent ywd.EntityId,
	pos y3d.Vec3,
	orientation y3d.Quaternion,
) {
	world.AddComponent(ent, ywd.MoveComponent, &Move{})
	world.AddComponent(ent, ywd.TransformComponent, NewTransform())
	world.AddComponent(ent, ywd.PositionComponent, pos)
	world.AddComponent(ent, ywd.OrientationComponent, orientation)
}

func AddEntityAsInstance(w *ywd.World, holder ywd.EntityId, instance ywd.EntityId) {
	if !w.HasComponent(holder, ywd.InstanceComponent) {
		instances := make([]ywd.EntityId, 0)
		instances = append(instances, instance)
		slices.Sort(instances)
		box := w.GetComponent(holder, ywd.BoundingBoxComponent).(yphy.Box)
		w.AddComponent(instance, ywd.BoundingBoxComponent, box)
		w.AddComponent(holder, ywd.InstanceComponent, instances)
		w.AddComponent(instance, ywd.InstanceOfComponent, holder)
		return
	}
	instances := w.GetComponent(holder, ywd.InstanceComponent).([]ywd.EntityId)
	instances = append(instances, instance)
	slices.Sort(instances)
	box := w.GetComponent(holder, ywd.BoundingBoxComponent).(yphy.Box)
	w.AddComponent(instance, ywd.BoundingBoxComponent, box)
	w.SetComponent(holder, ywd.InstanceComponent, instances)
	w.AddComponent(instance, ywd.InstanceOfComponent, holder)
}

func AddEntityTransformData(
	world *ywd.World,
	ent ywd.EntityId,
	transform Transform) {
	world.AddComponent(ent, ywd.TransformComponent, transform)
}
