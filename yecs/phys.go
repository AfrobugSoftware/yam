package yecs

import (
	"math"
	"runtime"
	"sync"
	"yam/y3d"
)

//temp real physics to be implemented according to cyclone by ian milligton

type Box struct {
	Local y3d.AABB
	World y3d.AABB
}

type CollisonInfo struct {
	Point        y3d.Vec3
	Normal       y3d.Vec3
	CollidedWith EntityId
}

type PhysicsSystem struct {
	Gravity float32
	Wg      sync.WaitGroup
}

func (p *PhysicsSystem) Init()     {}
func (p *PhysicsSystem) Shutdown() {}
func (p *PhysicsSystem) Query() []ComponentId {
	return []ComponentId{BoxComponent, MoveComponent, TransformComponent}
}

func (p *PhysicsSystem) SegmentCast(l *y3d.LineSegment, w *World, entites []EntityId) (CollisonInfo, bool) {
	collied := false
	closestT := float32(math.Inf(1))
	var outCol CollisonInfo
	for _, e := range entites {
		box := w.GetComponent(e, BoxComponent).(Box)
		t, intersect := l.IntersectAABB(box.World)
		if intersect {
			if t < closestT {
				closestT = t
				outCol.Point = l.PointOnLine(t)
				outCol.Normal = l.GetHitNormal(box.World, t)
				outCol.CollidedWith = e
				collied = true
			}
		}
	}
	return outCol, collied
}

func (p *PhysicsSystem) Run(w *World, dt float64, sub []EntityId, entites []EntityId) {
	for _, e := range sub {
		segmentLength := float32(30.0)
		transform := w.GetComponent(e, TransformComponent).(Transform)
		//move := w.GetComponent(e, MoveComponent).(Move)

		start := transform.Position
		dir := transform.GetForward()
		end := y3d.Add(start, y3d.Smul(dir, segmentLength))
		lineSegment := y3d.LineSegment{
			Start: start,
			End:   end,
		}
		info, intersect := p.SegmentCast(&lineSegment, w, entites)
		if intersect {
			dir = y3d.ReflectVec3(dir, info.Normal)
			(&transform).RotateToFoward(dir)
		}

		w.SetComponent(e, TransformComponent, transform)
	}
	p.Wg.Done()
}

func (p *PhysicsSystem) Update(w *World, dt float64, entites []EntityId) {
	brk := len(entites) / runtime.NumCPU()
	for i := range runtime.NumCPU() {
		p.Wg.Add(1)
		offset := i * brk
		go p.Run(w, dt, entites[offset:offset+brk], entites)
	}
	p.Wg.Wait()
}
