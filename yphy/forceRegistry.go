package yphy

import (
	"slices"
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
	for _, e := range r.Register {
		if w.HasComponent(e.Entity, ywd.ParticleComponent) {
			p := w.GetComponent(e.Entity, ywd.ParticleComponent).(Particle)
			p.AddForce(e.ForceGen, dt)
		}
	}
	//remove any deleted entity
	r.Register = slices.DeleteFunc(r.Register, func(r ForceRegisterEntry) bool {
		return !w.IsEntity(r.Entity)
	})
}
