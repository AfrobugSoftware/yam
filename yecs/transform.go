package yecs

import (
	"math"
	"runtime"
	"sync"
	"yam/y3d"
)

var (
	FORWARD = y3d.UNIT_Z
	UP      = y3d.UNIT_Y
	RIGHT   = y3d.UNIT_X
)

type Transform struct {
	Position       y3d.Vec3
	Rotation       y3d.Quaternion
	Scale          y3d.Vec3
	Transform      y3d.Mat4
	WorldTransform y3d.Mat4
	IsDirty        bool
}

func NewTransfromation() Transform {
	return Transform{
		Position:       y3d.Vec3{},
		Rotation:       y3d.IdenQuat(),
		Scale:          y3d.Vec3{X: 1.0, Y: 1.0, Z: 1.0},
		Transform:      y3d.Identity,
		WorldTransform: y3d.Identity,
		IsDirty:        false,
	}
}

func (trans *Transform) Recalulate() {
	scale := y3d.Scale(trans.Scale)
	rot := trans.Rotation.ToMat4()
	translation := y3d.Translation(trans.Position)
	trans.Transform = translation.Mul(rot).Mul(scale)
	trans.IsDirty = true
}

func (trans Transform) GetForward() y3d.Vec3 {
	return trans.Rotation.RotateVec3(FORWARD)
}

func (trans Transform) GetRight() y3d.Vec3 {
	return trans.Rotation.RotateVec3(RIGHT)
}

func (trans Transform) GetUp() y3d.Vec3 {
	return trans.Rotation.RotateVec3(UP)
}

func (t Transform) TransFormAABB(b y3d.AABB) y3d.AABB {
	(&b).Scale(t.Scale)
	//no rotation yet.
	(&b).Translate(t.Position)
	return b
}

func (t *Transform) RotateToFoward(forward y3d.Vec3) {
	dot := y3d.Dot(FORWARD, forward)
	angle := math.Acos(float64(dot))
	if dot > 0.9999 {
		t.Rotation = y3d.IdenQuat()
	} else if dot < -0.9999 {
		t.Rotation = y3d.FromAngleAxis(UP, math.Pi)
	} else {
		axis := y3d.Cross(FORWARD, forward)
		axis = y3d.Normalize(axis)
		t.Rotation = y3d.FromAngleAxis(axis, angle)
	}
	t.Recalulate()
}

type Hierarchy struct {
	Parent   EntityId
	Children []EntityId
}

type TransformSystem struct {
	Wg sync.WaitGroup
}

func (t *TransformSystem) Init()     {}
func (t *TransformSystem) Shutdown() {}
func (t *TransformSystem) Query() []ComponentId {
	return []ComponentId{HierarchyComponent, TransformComponent}
}

func (t *TransformSystem) ProceeEntity(w *World, e EntityId, parentTransform *Transform) {
	trans := w.GetComponent(e, TransformComponent).(Transform)
	h := w.GetComponent(e, HierarchyComponent).(Hierarchy)
	if trans.IsDirty {
		trans.WorldTransform = parentTransform.WorldTransform.Mul(trans.Transform)
		trans.IsDirty = false
		w.SetComponent(e, TransformComponent, trans)
	}
	for _, c := range h.Children {
		t.ProceeEntity(w, c, &trans)
	}
}

func (t *TransformSystem) Run(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		h := w.GetComponent(e, HierarchyComponent).(Hierarchy)
		if h.Parent == NullEntity {
			trans := w.GetComponent(e, TransformComponent).(Transform)
			trans.WorldTransform = trans.Transform
			for _, c := range h.Children {
				t.ProceeEntity(w, c, &trans)
			}
		}
	}
}

func (t *TransformSystem) Update(w *World, dt float64, entites []EntityId) {
	cpu := runtime.NumCPU()
	brk := len(entites) / cpu
	if brk < cpu {
		t.Wg.Add(1)
		go t.Run(w, dt, entites)
	} else {
		//does not work with odd entities
		for i := range cpu {
			t.Wg.Add(1)
			offset := i * brk
			if offset >= len(entites) {
				break
			}
			go t.Run(w, dt, entites[offset:offset+brk])
		}
	}
	t.Wg.Wait()
}
