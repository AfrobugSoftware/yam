package ycore

import (
	"sync"
	"yam/y3d"
	"yam/yphy"
	"yam/ywd"
)

var (
	wg sync.WaitGroup
)

func UpdateGS(w *ywd.World, owner ywd.EntityId, hierarchy ywd.Hierarchy) {
	transform := w.GetComponent(owner, ywd.TransformComponent).(Transform)
	box := w.GetComponent(owner, ywd.BoundingBoxComponent).(yphy.Box)
	if hierarchy.Parent == ywd.NullEntity {
		transform.World = transform.Local
	} else {
		ptransform := w.GetComponent(hierarchy.Parent, ywd.TransformComponent).(Transform)
		transform.World = ptransform.World.Mul(transform.Local)
		box.World = transform.TransformAABB(box.Local)
	}
	w.SetComponent(owner, ywd.TransformComponent, transform)
	for _, c := range hierarchy.Children {
		h := w.GetComponent(c, ywd.HierarchyComponent).(ywd.Hierarchy)
		UpdateGS(w, c, h)
		b := w.GetComponent(c, ywd.BoundingBoxComponent).(yphy.Box)
		box.World.Max = y3d.Max(box.World.Max, b.World.Max)
		box.World.Min = y3d.Min(box.World.Min, b.World.Min)
	}
	w.SetComponent(owner, ywd.BoundingBoxComponent, box)
}

func UpdateTranform(w *ywd.World) {
	entites := w.Query(
		ywd.PositionComponent,
		ywd.OrientationComponent,
		ywd.ScaleComponent,
		ywd.TransformComponent,
		ywd.BoundingBoxComponent,
	)
	for _, e := range entites {
		wg.Go(func() {
			transform := w.GetComponent(e, ywd.TransformComponent).(Transform)
			scale := w.GetComponent(e, ywd.ScaleComponent).(y3d.Vec3)
			position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
			orientation := w.GetComponent(e, ywd.OrientationComponent).(y3d.Quaternion)
			box := w.GetComponent(e, ywd.BoundingBoxComponent).(yphy.Box)

			transform.World = y3d.TRS(position,
				orientation,
				scale)
			box.World = transform.TransformAABB(box.Local)
			w.SetComponent(e, ywd.BoundingBoxComponent, box)
			w.SetComponent(e, ywd.TransformComponent, transform)
		})
	}
	wg.Wait()
}

func UpdateMovement(w *ywd.World, deltaTime float32) {
	entites := w.Query(ywd.MoveComponent,
		ywd.PositionComponent,
		ywd.OrientationComponent)
	for _, e := range entites {
		//change to go routine pool
		wg.Go(
			func() {
				movement := w.GetComponent(e, ywd.MoveComponent).(*Move)
				movement.Update(w, e, deltaTime)
			})
	}
	wg.Wait()
}

func RenderSceneGraph(w *ywd.World, r *RenderManager, root ywd.EntityId) {
	vm := r.VertextManager
	q := make([]ywd.EntityId, 0, 1000)
	q = append(q, root)
	for len(q) > 0 {
		f := q[0]
		q = q[1:]
		box := w.GetComponent(f, ywd.BoundingBoxComponent).(yphy.Box)
		c := r.GetFrustumPlanes()
		if !box.IsVisiable(c) {
			continue
		}
		if w.HasComponent(f, ywd.HierarchyComponent) {
			hierarchy := w.GetComponent(f, ywd.HierarchyComponent).(ywd.Hierarchy)
			q = append(q, hierarchy.Children...)
		}
		if w.HasComponent(f, ywd.RenderableComponent, ywd.TransformComponent) {
			renderable := w.GetComponent(f, ywd.RenderableComponent).(Renderable)
			v, err := vm.GetCache(renderable.VertexType, renderable.VertexCacheId)
			if err != nil {
				continue //no cache for this entity
			}
			transform := w.GetComponent(f, ywd.TransformComponent).(Transform)
			if w.HasComponent(f, ywd.InstanceComponent) {
				instances := r.World.GetComponent(f, ywd.InstanceComponent).([]ywd.EntityId)
				renderable.DrawCmd.InstanceCount = uint32(len(instances))
				mats := make([]y3d.Mat4, len(instances))
				rs := make([]PerObjectRenderState, len(instances))
				tags := make([]uint32, len(instances))
				for i, ie := range instances {
					mats[i] = r.World.GetComponent(ie, ywd.TransformComponent).(Transform).World
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
