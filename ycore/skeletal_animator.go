package ycore

import (
	"errors"
	"sync"
	"unsafe"
	"yam/y3d"
	"yam/ywd"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	SAMPLER_TRANSLATION = 1 << iota
	SAMPLER_SCALE
	SAMPLER_ROTATION
)

type Joint struct {
	Id       int
	Parent   int
	Children []int
	Name     string
	BindPose y3d.Mat4
	Local    y3d.Mat4
	World    y3d.Mat4
}

func NewJoint(
	id int,
	parent int,
	name string,
	bindPose y3d.Mat4,
) *Joint {
	j := &Joint{
		Id:       id,
		Parent:   parent,
		Name:     name,
		BindPose: bindPose,
		World:    y3d.Identity,
		Local:    y3d.Identity,
		Children: make([]int, 0),
	}
	return j
}

type Skeleton struct {
	Id     int
	Joints []*Joint
	Order  []int
	jbo    uint32
	q      []int
}

func NewSkeleton(id int, joints []*Joint) *Skeleton {
	s := &Skeleton{
		Id:     id,
		Joints: joints,
		q:      make([]int, 0),
	}
	var ubo uint32
	gl.CreateBuffers(1, &ubo)
	gl.NamedBufferStorage(ubo, int(unsafe.Sizeof(y3d.Mat4{}))*len(joints),
		nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	s.jbo = ubo
	return s
}

func (s *Skeleton) Update() error {
	ptr := gl.MapNamedBufferRange(
		s.jbo,
		0,
		int(unsafe.Sizeof(y3d.Mat4{}))*len(s.Joints),
		gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_RANGE_BIT,
	)
	if ptr == nil {
		return errors.New("cannot map joint buffer")
	}
	jointP := unsafe.Slice((*y3d.Mat4)(ptr), len(s.Joints))
	s.q = s.q[:0]
	for i, j := range s.Joints {
		if j.Parent == -1 {
			s.q = append(s.q, i)
		}
	}
	for len(s.q) > 0 {
		i := s.q[0]
		s.q = s.q[1:]
		s.q = append(s.q, s.Joints[i].Children...)
		if s.Joints[i].Parent == -1 {
			s.Joints[i].World = s.Joints[i].Local
		} else {
			s.Joints[i].World =
				s.Joints[s.Joints[i].Parent].World.Mul(s.Joints[i].Local)
		}
		jointP[i] = s.Joints[i].World.Mul(s.Joints[i].BindPose)
	}
	gl.UnmapNamedBuffer(s.jbo)
	return nil
}

func (s *Skeleton) Destory() {
	s.Joints = s.Joints[:0]
	gl.DeleteBuffers(1, &s.jbo)
}

type Track[T any] struct {
	Times        []float32
	Values       []T
	PlaybackTime float32
}

func (track *Track[T]) GetT(dt float32, repeat bool) (float32, int, int) {
	lastStamp := track.Times[len(track.Times)-1]
	track.PlaybackTime += dt
	if track.PlaybackTime >= lastStamp {
		if repeat {
			track.PlaybackTime = 0.0
		}
	}
	lastpos, nextpos := -1, -1
	for i := range track.Times {
		if track.Times[i] >= track.PlaybackTime {
			nextpos = i
			break
		}
		lastpos = i
	}
	var t float32
	if lastpos != -1 && nextpos != -1 {
		t = (track.PlaybackTime - track.Times[lastpos]) /
			(track.Times[nextpos] - track.Times[lastpos])
	}
	return t, lastpos, nextpos
}

type Sampler struct {
	Id           int
	PlaybackTime float32
	Flag         uint32
	Repeat       bool
	TimeStamps   []float32
	rotation     Track[y3d.Quaternion]
	translation  Track[y3d.Vec3]
	scale        Track[y3d.Vec3]

	OutRotation    y3d.Quaternion
	OutTranslation y3d.Vec3
	OutScale       y3d.Vec3
}

func (s *Sampler) Update(delta float32) {
	if s.Flag&SAMPLER_SCALE != 0 {
		t, lastpos, nextpos := s.scale.GetT(delta, s.Repeat)
		if lastpos != -1 && nextpos != -1 {
			s.OutScale = y3d.Lerp(s.scale.Values[lastpos], s.scale.Values[nextpos], t)
		} else if lastpos == -1 {
			s.OutScale = s.scale.Values[nextpos]
		} else {
			s.OutScale = s.scale.Values[lastpos]
		}
	}
	if s.Flag&SAMPLER_TRANSLATION != 0 {
		t, lastpos, nextpos := s.translation.GetT(delta, s.Repeat)
		if lastpos != -1 && nextpos != -1 {
			s.OutTranslation = y3d.Lerp(s.translation.Values[lastpos], s.translation.Values[nextpos], t)
		} else if lastpos == -1 {
			s.OutTranslation = s.translation.Values[nextpos]
		} else {
			s.OutTranslation = s.translation.Values[lastpos]
		}
	}
	if s.Flag&SAMPLER_ROTATION != 0 {
		t, lastpos, nextpos := s.rotation.GetT(delta, s.Repeat)
		if lastpos != -1 && nextpos != -1 {
			s.OutRotation = y3d.Slerp(s.rotation.Values[lastpos],
				s.rotation.Values[nextpos], t)
		} else if lastpos == -1 {
			s.OutRotation = s.rotation.Values[nextpos]
		} else {
			s.OutRotation = s.rotation.Values[lastpos]
		}
	}

}

type Animation struct {
	Samplers map[int]*Sampler
}

func NewAnimation() *Animation {
	return &Animation{
		Samplers: make(map[int]*Sampler),
	}
}

func (a *Animation) Play(delta float32) {
	for _, s := range a.Samplers {
		s.Update(delta)
	}
}

func (a *Animation) UpdateSkeleton(s *Skeleton) {
	for i := range s.Joints {
		sampler, ok := a.Samplers[s.Joints[i].Id]
		if ok {
			s.Joints[i].Local = y3d.TRS(sampler.OutTranslation, sampler.OutRotation, sampler.OutScale)
		}
	}

}

func UpdateAnimation(w *ywd.World, delta float32) {
	entites := w.Query(ywd.AnimationComponent)
	var wg sync.WaitGroup
	for _, e := range entites {
		wg.Go(func() {
			anims := w.GetComponent(e, ywd.AnimationComponent).([]*Animation)
			anim := anims[0]
			anim.Play(delta)
			if w.HasComponent(e, ywd.SkeletonComponent) {
				s := w.GetComponent(e, ywd.SkeletonComponent).(*Skeleton)
				anim.UpdateSkeleton(s)
				w.SetComponent(e, ywd.SkeletonComponent, s)
			}
			w.SetComponent(e, ywd.AnimationComponent, anims)
		})
	}
	wg.Wait()
	//testing to see if the error is because we are calling from a different thread
	//the skeleton is updated on the main thread because it is pushed to the gpu after update
	for _, e := range entites {
		if w.HasComponent(e, ywd.SkeletonComponent) {
			s := w.GetComponent(e, ywd.SkeletonComponent).(*Skeleton)
			err := s.Update()
			if err != nil {
				panic(err)
			}
			w.SetComponent(e, ywd.SkeletonComponent, s)
		}
	}
}

type Animator struct {
	Animations []*Animation
	Current    int
}

func NewAnimator(animations []*Animation) *Animator {
	return &Animator{
		Animations: animations,
		Current:    0,
	}
}
