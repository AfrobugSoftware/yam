package ycore

import (
	"slices"
	"yam/yphy"
)

type ForceRegisterEntry struct {
	Entity   EntityId
	ForceGen yphy.ForceGenerator
}

type ForceRegistry struct {
	Register []ForceRegisterEntry
}

func NewForceRegistry() *ForceRegistry {
	return &ForceRegistry{
		Register: make([]ForceRegisterEntry, 0),
	}
}

func (r *ForceRegistry) Update(w *World, dt float32) {
	for _, e := range r.Register {
		if w.HasComponent(e.Entity, ParticleComponent) {
			p := w.GetComponent(e.Entity, ParticleComponent).(yphy.Particle)
			p.AddForce(e.ForceGen, dt)
		}
	}
	//remove any deleted entity
	r.Register = slices.DeleteFunc(r.Register, func(r ForceRegisterEntry) bool {
		return !w.IsEntity(r.Entity)
	})
}
