package yecs

import (
	"log"
	"runtime"
	"sort"
	"sync"
	"time"
	"yam/y3d"
)

const (
	ROOT = iota
)

var (
	MatrixPalette           = map[int]y3d.Mat4{}
	AnimatedSpatialCompnent = RegisterComponent[AnimatedSpatial]()
)

type Joint struct {
	Id              int
	Parent          *Joint
	Children        []*Joint
	Position        y3d.Vec3
	Rotation        y3d.Quaternion
	GlobalTransform y3d.Mat4
	InvBindPose     y3d.Mat4
}

type Skeleton []Joint

type KeyFrame struct {
	Position y3d.Vec3
	Rotation y3d.Quaternion
}

type Animation3D struct {
	Name            string
	Duration        time.Duration
	FrameDuration   time.Duration
	KeyFrames       map[int][]KeyFrame //int is the joint Id, and the []KeyFrames are the key frames or tracks of the joint
	CurrentSkeleton Skeleton
	NumJoints       int
}

type AnimatedSpatial struct {
	Spatial
	Animation     *Animation3D
	PlayRate      float32
	AnimationTime time.Duration
	PoseCache     []float32
}

func (j *Joint) PropagateToLeaf() {
	for _, i := range j.Children {
		if i != nil {
			i.GlobalTransform = j.GlobalTransform.Mul(i.GetLocal())
			i.PropagateToLeaf()
		}
	}
}

func (j *Joint) UpdateTransform(position y3d.Vec3, rotation y3d.Quaternion) {
	j.Position = position
	j.Rotation = rotation
}

func (j *Joint) GetLocal() y3d.Mat4 {
	local := y3d.Translation(j.Position)
	local = local.Mul(j.Rotation.ToMat4())
	return local
}

func (j *Joint) CalculateInverseBindPose() {
	local := y3d.Translation(j.Position)
	local = local.Mul(j.Rotation.ToMat4())
	if j.Parent != nil {
		j.GlobalTransform = j.Parent.GlobalTransform.Mul(local)
	} else {
		j.GlobalTransform = local
	}
	j.InvBindPose = j.GlobalTransform
	(&j.InvBindPose).Invert()
}

func (j *Joint) ToMat4() y3d.Mat4 {
	return j.GlobalTransform.Mul(j.InvBindPose)
}

func Interpolate(a, b *KeyFrame, t float32) KeyFrame {
	return KeyFrame{
		Rotation: y3d.Slerp(a.Rotation, b.Rotation, float64(t)),
		Position: y3d.Lerp(a.Position, b.Position, t),
	}
}

func RegisterPalette(skeleton Skeleton) {
	for _, s := range skeleton {
		MatrixPalette[s.Id] = s.ToMat4()
	}
}

func GatherMatrix(skeleton Skeleton) []float32 {
	flat := make([]float32, len(skeleton)*16)
	for _, s := range skeleton {
		m := s.ToMat4()
		copy(flat[s.Id*16:], m[:])
	}
	return flat
}

func (a *Animation3D) SortJoints() {
	sort.Slice(a.CurrentSkeleton, func(i, j int) bool {
		return a.CurrentSkeleton[i].Id < a.CurrentSkeleton[j].Id
	})
}

func (a *Animation3D) GetPose(inTime time.Duration) []float32 {
	if len(a.CurrentSkeleton) == 0 {
		log.Printf("no skeleton for: %s\n", a.Name)
		return nil
	}
	fd := a.FrameDuration.Seconds()
	it := inTime.Seconds()
	frame := int(it / fd)
	nextFrame := frame + 1
	pct := it / (fd - float64(frame))
	for i := range a.CurrentSkeleton {
		kf, ok := a.KeyFrames[i]
		if ok {
			f := kf[frame]
			f2 := kf[nextFrame]
			pf := Interpolate(&f, &f2, float32(pct))
			(&a.CurrentSkeleton[i]).UpdateTransform(pf.Position, pf.Rotation)
		} else {
			(&a.CurrentSkeleton[i]).UpdateTransform(y3d.Vec3{}, y3d.IdenQuat())
		}
	}
	a.CurrentSkeleton[ROOT].GlobalTransform = (&a.CurrentSkeleton[ROOT]).GetLocal()
	a.CurrentSkeleton[ROOT].PropagateToLeaf()
	return GatherMatrix(a.CurrentSkeleton)
}

type Animation3dSystem struct {
	Wg sync.WaitGroup
}

func (as *Animation3dSystem) Init()     {}
func (as *Animation3dSystem) Shutdown() {}
func (as *Animation3dSystem) Query() []ComponentId {
	return []ComponentId{AnimatedSpatialCompnent}
}

func (as *Animation3dSystem) Run(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		as := w.GetComponent(e, AnimatedSpatialCompnent).(AnimatedSpatial)
		as.AnimationTime += time.Duration((dt * float64(as.PlayRate)) * float64(time.Second))
		for as.AnimationTime > as.Animation.Duration {
			as.AnimationTime -= as.Animation.Duration
		}

		as.PoseCache = as.Animation.GetPose(as.AnimationTime)

		w.SetComponent(e, AnimatedSpatialCompnent, as)
	}
}

func (as *Animation3dSystem) Update(w *World, dt float64, entites []EntityId) {
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
