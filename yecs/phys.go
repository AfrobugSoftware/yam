package yecs

import (
	"math"
	"runtime"
	"sort"
	"sync"
	"yam/y3d"
)

//temp real physics to be implemented according to cyclone by ian milligton

type Box struct {
	Local y3d.AABB
	World y3d.AABB
	F     func(a, b EntityId, w *World)
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

func (p *PhysicsSystem) RemoveInterprenetrationFromA(a, b y3d.AABB, aPos y3d.Vec3) y3d.Vec3 {
	dx1 := b.Min.X - a.Max.X
	dx2 := b.Max.X - a.Min.X
	dy1 := b.Min.Y - a.Max.Y
	dy2 := b.Max.Y - a.Min.Y
	dz1 := b.Min.Z - a.Max.Z
	dz2 := b.Max.Z - a.Min.Z

	var dx, dy, dz float64
	if math.Abs(float64(dx1)) < math.Abs(float64(dx2)) {
		dx = float64(dx1)
	} else {
		dx = float64(dx2)
	}
	if math.Abs(float64(dy1)) < math.Abs(float64(dy2)) {
		dy = float64(dy1)
	} else {
		dy = float64(dy2)
	}
	if math.Abs(float64(dz1)) < math.Abs(float64(dz2)) {
		dz = float64(dz1)
	} else {
		dz = float64(dz2)
	}

	if math.Abs(dx) <= math.Abs(dy) && math.Abs(dx) <= math.Abs(dz) {
		aPos.X += float32(dx)
	} else if math.Abs(dy) <= math.Abs(dx) && math.Abs(dy) <= math.Abs(dz) {
		aPos.Y += float32(dy)
	} else {
		aPos.Z += float32(dz)
	}
	return aPos
}

func (p *PhysicsSystem) RemoveInterprenetrationFromA2D(a, b y3d.AABB, aPos y3d.Vec3) y3d.Vec3 {
	dx1 := b.Min.X - a.Max.X
	dx2 := b.Max.X - a.Min.X
	dy1 := b.Min.Y - a.Max.Y
	dy2 := b.Max.Y - a.Min.Y

	var dx, dy float64
	if math.Abs(float64(dx1)) < math.Abs(float64(dx2)) {
		dx = float64(dx1)
	} else {
		dx = float64(dx2)
	}
	if math.Abs(float64(dy1)) < math.Abs(float64(dy2)) {
		dy = float64(dy1)
	} else {
		dy = float64(dy2)
	}

	if math.Abs(dx) <= math.Abs(dy) {
		aPos.X += float32(dx)
	} else if math.Abs(dy) <= math.Abs(dx) {
		aPos.Y += float32(dy)
	}
	return aPos
}

func (p *PhysicsSystem) TestSweepAndPrune(w *World, entites []EntityId) {
	type Ebox struct {
		B Box
		E EntityId
	}
	boxes := make([]Ebox, 0)
	for _, e := range entites {
		box := w.GetComponent(e, BoxComponent).(Box)
		boxes = append(boxes, Ebox{B: box, E: e})
	}
	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i].B.World.Min.X < boxes[j].B.World.Min.X
	})
	size := len(boxes)
	for i := range size {
		a := &boxes[i]
		max := a.B.World.Max.X
		for j := i + 1; j < size; j++ {
			b := &boxes[j]
			if b.B.World.Min.X > max {
				break
			} else {
				if a.B.World.Overlaps(b.B.World) {
					//how can I handle
					if a.B.F != nil {
						a.B.F(a.E, b.E, w)
					}
				}
			}
		}
	}
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
	cpu := runtime.NumCPU()
	brk := len(entites) / cpu
	if brk < cpu {
		p.Wg.Add(1)
		go p.Run(w, dt, entites, entites)
	} else {
		//does not work with odd entities
		for i := range cpu {
			p.Wg.Add(1)
			offset := i * brk
			if offset >= len(entites) {
				break
			}
			go p.Run(w, dt, entites[offset:offset+brk], entites)
		}
	}
	p.Wg.Wait()
}
