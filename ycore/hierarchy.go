package ycore

import "yam/y3d"

type Hierarchy struct {
	Children []EntityId
	Parent   EntityId
}

func UpdateGS(w *World, owner EntityId, hierarchy Hierarchy) {
	transform := w.GetComponent(owner, TransformComponent).(Transform)
	box := w.GetComponent(owner, BoundingBoxComponent).(Box)
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
		UpdateGS(w, c, h)
		b := w.GetComponent(c, BoundingBoxComponent).(Box)
		box.World.Max = y3d.Max(box.World.Max, b.World.Max)
		box.World.Min = y3d.Min(box.World.Min, b.World.Min)
	}
	w.SetComponent(owner, BoundingBoxComponent, box)
}

func RenderScene(r *RenderManager, w *World, root EntityId) {
	vm := r.VertextManager
	q := make([]EntityId, 0, 1000)
	q = append(q, root)
	for len(q) > 0 {
		f := q[0]
		q = q[1:]
		box := w.GetComponent(f, BoundingBoxComponent).(Box)
		c := r.GetFrustum()
		if box.World.Cull(c[:]) {
			continue
		}
		hierarchy := w.GetComponent(f, HierarchyComponent).(Hierarchy)
		q = append(q, hierarchy.Children...)
		if w.HasComponent(f, RenderableComponent, TransformComponent) {
			renderable := w.GetComponent(f, RenderableComponent).(Renderable)
			v, err := vm.GetCache(renderable.VertexType, renderable.SkinId)
			if err != nil {
				continue //no cache for this entity
			}
			transform := w.GetComponent(f, TransformComponent).(Transform)
			if w.HasComponent(f, MatInstanceComponent, RenderParamsInstanceComponent) {
				imat := w.GetComponent(f, MatInstanceComponent).([]y3d.Mat4)
				ip := w.GetComponent(f, RenderParamsInstanceComponent).([]PerObjectRenderState)
				renderable.DrawCmd.InstanceCount = uint32(len(imat))
				v.AddInstances(&renderable.DrawCmd)
				v.AddRenderState(int(renderable.DrawCmd.BaseInstance), ip...)
				v.AddMatrix(int(renderable.DrawCmd.BaseInstance), imat...)
			} else {
				v.AddInstances(&renderable.DrawCmd)
				v.AddMatrix(int(renderable.DrawCmd.BaseInstance), transform.World) //this needs to be world
				v.AddRenderState(int(renderable.DrawCmd.BaseInstance), renderable.RenderParams)
			}
			v.AddDrawCommand(renderable.DrawCmd)
		}
	}
}
