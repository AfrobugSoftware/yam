package yecs

import (
	"runtime"
	"sync"
	"yam/y3d"
)

const (
	ROOT = iota
)

type AnimTarget uint8

const (
	AnimTargetTranslation AnimTarget = 0x01
	AnimTargetRotation    AnimTarget = 0x02
	AnimTargetScale       AnimTarget = 0x04
)

type KeyFrame struct {
	Target   AnimTarget
	Position y3d.Vec3
	Scale    y3d.Vec3
	Rotation y3d.Quaternion
}

type TimeStamps []float32
type KeyFrames []KeyFrame

func GetTimeFromStamps(max, min, current float32) float32 {
	return (current - min) / (max - min)
}

func Interpolate(a, b *KeyFrame, t float32) KeyFrame {
	kf := KeyFrame{}
	if a.Target&AnimTargetTranslation != 0 {
		kf.Position = y3d.Lerp(a.Position, b.Position, t)
	}
	if a.Target&AnimTargetRotation != 0 {
		kf.Rotation = y3d.Slerp(a.Rotation, b.Rotation, float64(t))
	}
	if a.Target&AnimTargetScale != 0 {
		kf.Scale = y3d.Lerp(a.Scale, b.Scale, t)
	}
	kf.Target = a.Target
	return kf
}

func SetTransform(a *KeyFrame, b *Transform) {
	if a.Target&AnimTargetTranslation != 0 {
		b.Position = a.Position
	}
	if a.Target&AnimTargetRotation != 0 {
		b.Rotation = a.Rotation
	}
	if a.Target&AnimTargetScale != 0 {
		b.Scale = a.Scale
	}
	b.Recalulate()
}

type Animation struct {
	Id              int
	Duration        float32
	KeyFrames       KeyFrames
	TimeStamps      TimeStamps
	InverseBindPose *y3d.Mat4
}

func CalculateInverseBindPose(w *World, e EntityId, parentWorld y3d.Mat4) {
	h := w.GetComponent(e, HierarchyComponent).(Hierarchy)
	t := w.GetComponent(e, TransformComponent).(Transform)
	a := w.GetComponent(e, AnimationComponent).(Animation)

	t.World = parentWorld.Mul(t.Local)
	a.InverseBindPose = &y3d.Mat4{}
	*(a.InverseBindPose) = t.World
	a.InverseBindPose.Invert()

	w.SetComponent(e, TransformComponent, t)
	w.SetComponent(e, AnimationComponent, a)
	for _, i := range h.Children {
		CalculateInverseBindPose(w, i, t.World)
	}
}

// skeleton is sorted by Id
func GatherMatrix(skeleton []y3d.Mat4) []float32 {
	flat := make([]float32, len(skeleton)*16)
	for i, m := range skeleton {
		copy(flat[i*16:], m[:])
	}
	return flat
}

type AnimationSystem struct {
	Wg sync.WaitGroup
}

func (as *AnimationSystem) Init()     {}
func (as *AnimationSystem) Shutdown() {}
func (as *AnimationSystem) Query() []ComponentId {
	return []ComponentId{AnimationComponent, TransformComponent}
}

func (as *AnimationSystem) Run(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		t := w.GetComponent(e, TransformComponent).(Transform)
		a := w.GetComponent(e, AnimationComponent).(Animation)

		var i int
		a.Duration += float32(dt)
		var found bool
		for i = range a.TimeStamps {
			if a.Duration <= a.TimeStamps[i] {
				found = true
				break
			}
		}
		var kf KeyFrame
		if i == 0 || !found {
			//reached the end
			a.Duration = 0
			kf = a.KeyFrames[0]
		} else {
			if i >= len(a.KeyFrames) || i < 0 ||
				i-1 < 0 {
				continue //not sure what do here
			}
			frameB := a.KeyFrames[i]
			frameA := a.KeyFrames[i-1]
			timeB := a.TimeStamps[i]
			timeA := a.TimeStamps[i-1]

			pct := GetTimeFromStamps(timeB, timeA, a.Duration)
			kf = Interpolate(&frameA, &frameB, pct)
		}

		SetTransform(&kf, &t)
		w.SetComponent(e, TransformComponent, t)
		w.SetComponent(e, AnimationComponent, a)
	}
}

func (as *AnimationSystem) Update(w *World, dt float64, entites []EntityId) {
	cpu := runtime.NumCPU()
	brk := len(entites) / cpu
	if brk < cpu {
		as.Wg.Add(1)
		go as.Run(w, dt, entites)
	} else {
		for i := range cpu {
			as.Wg.Add(1)
			offset := i * brk
			if offset >= len(entites) {
				break
			}
			go as.Run(w, dt, entites[offset:offset+brk])
		}
	}
	as.Wg.Wait()
}
