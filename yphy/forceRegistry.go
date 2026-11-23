package yphy

import (
	"slices"
	"yam/y3d"
	"yam/ywd"
)

type ForceRegisterEntry struct {
	Entity   ywd.EntityId
	ForceGen ForceGenerator
}

type ForceRegistry struct {
	Register []ForceRegisterEntry
}

func NewForceRegistry() *ForceRegistry {
	return &ForceRegistry{
		Register: make([]ForceRegisterEntry, 0),
	}
}

func (r *ForceRegistry) Add(e ywd.EntityId, f ForceGenerator) {
	r.Register = append(r.Register, ForceRegisterEntry{
		Entity:   e,
		ForceGen: f,
	})
}

func (r *ForceRegistry) Remove(e ywd.EntityId, f ForceGenerator) {
	r.Register = slices.DeleteFunc(r.Register, func(r ForceRegisterEntry) bool {
		return r.Entity == e && r.ForceGen == f
	})
}

func (r *ForceRegistry) Update(w *ywd.World, dt float32) {
	r.Register = slices.DeleteFunc(r.Register, func(r ForceRegisterEntry) bool {
		return !w.IsEntity(r.Entity)
	})
	for _, e := range r.Register {
		if w.HasComponent(e.Entity, ywd.ForceComponent) {
			force := w.GetComponent(e.Entity, ywd.ForceComponent).(y3d.Vec3)
			newForce := e.ForceGen.Force(w, e.Entity, dt)
			force = y3d.Add(force, newForce)
			w.SetComponent(e.Entity, ywd.ForceComponent, force)
		}
	}
}
