package ycore

import (
	"sync"
	"yam/y3d"
	"yam/ywd"
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
	JointP []y3d.Mat4
	Order  []int
	q      []int
}

func NewSkeleton(id int, joints []*Joint) *Skeleton {
	s := &Skeleton{
		Id:     id,
		Joints: joints,
		JointP: make([]y3d.Mat4, len(joints)),
		q:      make([]int, 0),
	}

	return s
}

func (s *Skeleton) Update() error {
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
		s.JointP[i] = s.Joints[i].World.Mul(s.Joints[i].BindPose)
	}
	return nil
}

func (s *Skeleton) Destory() {
	s.Joints = s.Joints[:0]
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

// smaplers get blended to product local coordinate systems
func UpdateAnimation(w *ywd.World, delta float32) {
	entites := w.Query(ywd.AnimationComponent)
	var wg sync.WaitGroup
	for _, e := range entites {
		wg.Go(func() {
			anims := w.GetComponent(e, ywd.AnimationComponent).(*Animator)
			anim := anims.Animations[anims.Current]
			anim.Play(delta)
			if w.HasComponent(e, ywd.SkeletonComponent) {
				s := w.GetComponent(e, ywd.SkeletonComponent).(*Skeleton)
				anim.UpdateSkeleton(s)
				err := s.Update()
				if err != nil {
					panic(err)
				}
				w.SetComponent(e, ywd.SkeletonComponent, s)
			}
			w.SetComponent(e, ywd.AnimationComponent, anims)
		})
	}
	wg.Wait()
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
