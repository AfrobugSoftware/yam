package ycore

import (
	"errors"
	"sync"
	"unsafe"
	"yam/y3d"

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
	Name     string
	BindPose y3d.Mat4
	Local    y3d.Mat4
	World    y3d.Mat4
}

func NewJoint(
	id int,
	parent int,
	name string,
	pos y3d.Vec3,
	rot y3d.Quaternion,
	scale y3d.Vec3,
	undeformedMatrix y3d.Mat4,
	bindPose y3d.Mat4,
) *Joint {
	j := &Joint{
		Id:       id,
		Parent:   parent,
		Name:     name,
		BindPose: bindPose,
	}
	j.Local = y3d.TRS(pos, rot, scale).Mul(undeformedMatrix)
	return j
}

type Skeleton struct {
	Id     int
	Joints []*Joint
	jbo    uint32
}

func NewSkeleton(id int, joints []*Joint) *Skeleton {
	s := &Skeleton{
		Id:     id,
		Joints: joints,
	}
	var ubo uint32
	gl.CreateBuffers(1, &ubo)
	gl.NamedBufferStorage(ubo, int(unsafe.Sizeof(y3d.Mat4{}))*len(joints),
		nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	s.jbo = ubo
	return s
}

func (s *Skeleton) Update() error {
	//joint parents come before the children
	ptr := gl.MapNamedBufferRange(
		s.jbo,
		0,
		int(unsafe.Sizeof(y3d.Mat4{}))*len(s.Joints),
		gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_RANGE_BIT,
	)
	jointP := unsafe.Slice((*y3d.Mat4)(ptr), len(s.Joints))
	for i := range s.Joints {
		if s.Joints[i].Parent == -1 {
			s.Joints[i].World =
				s.Joints[i].Local
		} else {
			p := s.Joints[s.Joints[i].Parent]
			s.Joints[i].World =
				p.World.Mul(s.Joints[i].Local)
		}
		jointP[i] = s.Joints[i].World.Mul(s.Joints[i].BindPose)
	}
	if ptr == nil {
		return errors.New("cannot map joint buffer")
	}
	gl.UnmapNamedBuffer(s.jbo)
	return nil
}

func (s *Skeleton) Destory() {
	s.Joints = s.Joints[:0]
	gl.DeleteBuffers(1, &s.jbo)
}

type Sampler struct {
	Id           int
	PlaybackTime float32
	Flag         uint32
	Repeat       bool
	TimeStamps   []float32
	rotation     []y3d.Quaternion
	translation  []y3d.Vec3
	scale        []y3d.Vec3

	OutRotation    y3d.Quaternion
	OutTranslation y3d.Vec3
	OutScale       y3d.Vec3
}

func (s *Sampler) Update(delta float32) {
	s.PlaybackTime += delta
	lastStamp := s.TimeStamps[len(s.TimeStamps)-1]
	if s.PlaybackTime > lastStamp {
		if s.Repeat {
			s.PlaybackTime = 0.0
		} else {
			return
		}
	}
	lastpos, nextpos := -1, -1
	for i := range s.TimeStamps {
		if s.TimeStamps[i] >= s.PlaybackTime {
			nextpos = i
			break
		}
		lastpos = i
	}
	if lastpos != -1 && nextpos != -1 {
		t := (s.PlaybackTime - s.TimeStamps[lastpos]) /
			(s.TimeStamps[nextpos] - s.TimeStamps[lastpos])
		if s.Flag&SAMPLER_SCALE != 0 {
			s.OutScale = y3d.Lerp(s.scale[lastpos], s.scale[nextpos], t)
		}
		if s.Flag&SAMPLER_TRANSLATION != 0 {
			s.OutTranslation = y3d.Lerp(s.translation[lastpos],
				s.translation[nextpos], t)
		}
		if s.Flag&SAMPLER_ROTATION != 0 {
			s.OutRotation = y3d.Slerp(s.rotation[lastpos],
				s.rotation[nextpos], t)
		}
	} else if lastpos == -1 {
		if s.Flag&SAMPLER_SCALE != 0 {
			s.OutScale = s.scale[nextpos]
		}
		if s.Flag&SAMPLER_TRANSLATION != 0 {
			s.OutTranslation = s.translation[nextpos]
		}
		if s.Flag&SAMPLER_ROTATION != 0 {
			s.OutRotation = s.rotation[nextpos]
		}
	} else {
		if s.Flag&SAMPLER_SCALE != 0 {
			s.OutScale = s.scale[lastpos]
		}
		if s.Flag&SAMPLER_TRANSLATION != 0 {
			s.OutTranslation = s.translation[lastpos]
		}
		if s.Flag&SAMPLER_ROTATION != 0 {
			s.OutRotation = s.rotation[lastpos]
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

func UpdateEntityAnimation(w *World, delta float32) {
	entites := w.Query(AnimationComponent)
	var wg sync.WaitGroup
	for _, e := range entites {
		wg.Go(func() {
			anim := w.GetComponent(e, AnimationComponent).(*Animation)
			anim.Play(delta)
			if w.HasComponent(e, SkeletonComponent) {
				s := w.GetComponent(e, SkeletonComponent).(*Skeleton)
				anim.UpdateSkeleton(s)
				w.SetComponent(e, SkeletonComponent, s)
			}
			w.SetComponent(e, AnimationComponent, anim)
		})
	}
	wg.Wait()
}
