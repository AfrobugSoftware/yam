package ycore

import (
	"sync"
	"yam/y3d"
	"yam/ycontroller"
	"yam/yphy"
)

type Hierarchy struct {
	Children []EntityId
	Parent   EntityId
}

func (w *World) UpdateGS(owner EntityId, hierarchy Hierarchy) {
	transform := w.GetComponent(owner, TransformComponent).(Transform)
	box := w.GetComponent(owner, BoundingBoxComponent).(yphy.Box)
	if hierarchy.Parent == NullEntity {
		transform.World = transform.Local
	} else {
		ptransform := w.GetComponent(hierarchy.Parent, TransformComponent).(Transform)
		transform.World = ptransform.World.Mul(transform.Local)
		box.World = transform.TransformAABB(box.Local)
	}
	w.SetComponent(owner, TransformComponent, transform)
	for _, c := range hierarchy.Children {
		h := w.GetComponent(c, HierarchyComponent).(Hierarchy)
		w.UpdateGS(c, h)
		b := w.GetComponent(c, BoundingBoxComponent).(yphy.Box)
		box.World.Max = y3d.Max(box.World.Max, b.World.Max)
		box.World.Min = y3d.Min(box.World.Min, b.World.Min)
	}
	w.SetComponent(owner, BoundingBoxComponent, box)
}

func (w *World) UpdateMovement(deltaTime float32) {
	entites := w.Query(
		TransformComponent,
		ScaleComponent,
		MoveControllerComponent)
	var wg sync.WaitGroup
	for _, e := range entites {
		//change to go routine pool
		wg.Go(
			func() {
				movement := w.GetComponent(e, MoveControllerComponent).(*ycontroller.MovementController)
				transform := w.GetComponent(e, TransformComponent).(Transform)
				movement.Update(deltaTime)
				scaleFactor := float32(1.0) //l.World.GetComponent(e, ycore.ScaleComponent).(float32)
				transform.Local = y3d.TRS(movement.Position, movement.Orientation, y3d.Vec3{
					X: scaleFactor,
					Y: scaleFactor,
					Z: scaleFactor,
				})
				w.SetComponent(e, TransformComponent, transform)
			})
	}
	wg.Wait()
	w.UpdateGS(w.SceneRoot, w.SceneRootHierarchy)
	UpdateEntityAnimation(w, deltaTime)
	w.BoundingTree.BuildTree(entites)
}

func (w *World) RenderSceneGraph(r *RenderManager, root EntityId) {
	vm := r.VertextManager
	q := make([]EntityId, 0, 1000)
	q = append(q, root)
	for len(q) > 0 {
		f := q[0]
		q = q[1:]
		box := w.GetComponent(f, BoundingBoxComponent).(yphy.Box)
		c := r.GetFrustum()
		if box.World.Cull(c[:]) {
			continue
		}
		if w.HasComponent(f, HierarchyComponent) {
			hierarchy := w.GetComponent(f, HierarchyComponent).(Hierarchy)
			q = append(q, hierarchy.Children...)
		}
		if w.HasComponent(f, RenderableComponent, TransformComponent) {
			renderable := w.GetComponent(f, RenderableComponent).(Renderable)
			v, err := vm.GetCache(renderable.VertexType, renderable.SkinId)
			if err != nil {
				continue //no cache for this entity
			}
			transform := w.GetComponent(f, TransformComponent).(Transform)
			if w.HasComponent(f, InstanceComponent) {
				instances := r.World.GetComponent(f, InstanceComponent).([]EntityId)
				renderable.DrawCmd.InstanceCount = uint32(len(instances))
				mats := make([]y3d.Mat4, len(instances))
				rs := make([]PerObjectRenderState, len(instances))
				tags := make([]uint32, len(instances))
				for i, ie := range instances {
					mats[i] = r.World.GetComponent(ie, TransformComponent).(Transform).World
					rs[i] = renderable.RenderParams
					tags[i] = uint32(ie)
				}
				v.AddInstances(&renderable.DrawCmd)
				v.AddRenderState(int(renderable.DrawCmd.BaseInstance), rs...)
				v.AddMatrix(int(renderable.DrawCmd.BaseInstance), mats...)
				v.AddTag(int(renderable.DrawCmd.BaseInstance), tags...)
			} else {
				v.AddInstances(&renderable.DrawCmd)
				v.AddMatrix(int(renderable.DrawCmd.BaseInstance), transform.World) //this needs to be world
				v.AddRenderState(int(renderable.DrawCmd.BaseInstance), renderable.RenderParams)
				v.AddTag(int(renderable.DrawCmd.BaseInstance), uint32(f))
			}
			v.AddDrawCommand(renderable.DrawCmd)
		}
	}
}
