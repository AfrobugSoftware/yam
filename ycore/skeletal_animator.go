package ycore

import (
	"errors"
	"unsafe"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	NIL_JOINT_PARENT = -1
	NO_ANIMATION     = -1
	ROOT             = 0 //root is alway at zero
)

type Animation struct {
	Name         string
	StartFrame   float32
	EndFrame     float32
	FPS          float32
	PlaybackTime float32
	IsActive     bool
	RunOnce      bool
	IsComplete   bool
	Root         int
	Joints       []*Joint
}

func NewAnimation(
	name string,
	fps float32,
	starframe, endframe float32,
	runOnce bool,
	root int,
	joints []*Joint,
) *Animation {
	return &Animation{
		Name:       name,
		StartFrame: starframe,
		EndFrame:   endframe,
		FPS:        fps,
		IsActive:   false,
		RunOnce:    runOnce,
		IsComplete: false,
		Joints:     joints,
		Root:       root,
	}
}

type TranslationSampler struct {
	Time float32
	Pos  y3d.Vec3
}

type RotationSampler struct {
	Time float32
	Rot  y3d.Quaternion
}

type ScaleSampler struct {
	Time  float32
	Scale y3d.Vec3
}

type Joint struct {
	Id         int
	Parent     int
	Name       string
	Children   []int
	KPos       []TranslationSampler
	KRot       []RotationSampler
	BindPose   y3d.Mat4
	Transform  *Transform
	IsAnimated bool
}

type Skeleton struct {
	Joints []*Joint
	jbo    uint32
}

func NewSkeleton(joints []*Joint) *Skeleton {
	s := &Skeleton{
		Joints: joints,
	}
	var ubo uint32
	gl.CreateBuffers(1, &ubo)
	gl.NamedBufferStorage(ubo, int(unsafe.Sizeof(y3d.Mat4{}))*len(joints),
		nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	s.jbo = ubo
	return s
}

func (s *Skeleton) Load() error {
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
	for i, j := range s.Joints {
		jointP[i] = j.Transform.World.Mul(j.BindPose)
	}
	gl.UnmapNamedBuffer(s.jbo)
	return nil
}

type SkeletalAnimator struct {
	Animation         []*Animation
	JointBufferObject uint32
	MaxJoints         int
	CurrentAnimation  int
}

func NewSkeletalAnimator(
	animations []*Animation,
	maxJoints int,
) *SkeletalAnimator {
	sa := &SkeletalAnimator{
		Animation:        animations,
		MaxJoints:        maxJoints,
		CurrentAnimation: NO_ANIMATION,
	}
	gl.CreateBuffers(1, &sa.JointBufferObject)
	gl.NamedBufferStorage(sa.JointBufferObject, int(unsafe.Sizeof(y3d.Mat4{}))*maxJoints, nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	return sa
}

func (sa *SkeletalAnimator) Add(animation *Animation, calcInverseBind bool) (int, error) {
	if len(animation.Joints) >= sa.MaxJoints {
		return NO_ANIMATION, errors.New("more joints in the animation that is expected")
	}
	sa.Animation = append(sa.Animation, animation)
	sa.CurrentAnimation = len(sa.Animation) - 1
	return sa.CurrentAnimation, nil
}

func (sa *SkeletalAnimator) SetAnimation(i int) error {
	if i >= len(sa.Animation) {
		return errors.New("invalid animation id")
	}
	sa.CurrentAnimation = i
	animation := sa.Animation[sa.CurrentAnimation]
	animation.PlaybackTime = 0
	animation.IsComplete = false
	animation.IsActive = true
	return nil
}

func (sa *SkeletalAnimator) Play(deltaTime float32) {
	if len(sa.Animation) == 0 || sa.CurrentAnimation == NO_ANIMATION {
		return
	}
	animation := sa.Animation[sa.CurrentAnimation]
	if animation.RunOnce && animation.IsComplete {
		return
	}
	animation.PlaybackTime += deltaTime
	if animation.PlaybackTime >= animation.EndFrame {
		animation.IsComplete = true
		animation.PlaybackTime = animation.StartFrame
	} else {
		animation.IsComplete = false
		//I do not think this is correct, my idea on the way the frames are calculated might be wrong
		//need data for test

		for _, j := range animation.Joints {
			if len(j.KPos) != 0 && len(j.KRot) != 0 && j.IsAnimated {
				lastpos, nextpos := -1, -1
				for i := range j.KPos {
					if j.KPos[i].Time >= animation.PlaybackTime {
						nextpos = i
						break
					}
					lastpos = i
				}
				if lastpos != -1 && nextpos != -1 {
					t := (animation.PlaybackTime - j.KPos[lastpos].Time) /
						(j.KPos[nextpos].Time - j.KPos[lastpos].Time)
					j.Transform.Position = y3d.Lerp(j.KPos[lastpos].Pos, j.KPos[nextpos].Pos, t)
				} else if lastpos == -1 {
					j.Transform.Position = j.KPos[nextpos].Pos
				} else {
					j.Transform.Position = j.KPos[lastpos].Pos
				}

				lastpos, nextpos = -1, -1
				for i := range j.KRot {
					if j.KRot[i].Time >= animation.PlaybackTime {
						nextpos = i
						break
					}
					lastpos = i
				}
				if lastpos != -1 && nextpos != -1 {
					t := (animation.PlaybackTime - j.KPos[lastpos].Time) /
						(j.KRot[nextpos].Time - j.KRot[lastpos].Time)
					j.Transform.Rotation = y3d.Slerp(j.KRot[lastpos].Rot, j.KRot[nextpos].Rot, float64(t))
				} else if lastpos == -1 {
					j.Transform.Rotation = j.KRot[nextpos].Rot
				} else {
					j.Transform.Rotation = j.KRot[lastpos].Rot
				}
				j.Transform.RecalulateNoScale()
				j.Transform.Local = j.Transform.Local.Mul(j.BindPose) // i think lol
			} else {
				j.Transform.Local = j.BindPose //copy the bind pos
			}
			//walk the tree from the root
			animation.Joints[animation.Root].Update(animation.Joints)
			sa.LoadBuffer() //upload data to the gpu
		}
	}
}

func (sa *SkeletalAnimator) Destory() {
	gl.DeleteBuffers(1, &sa.JointBufferObject)
	clear(sa.Animation)
}

func (sa *SkeletalAnimator) LoadBuffer() error {
	if sa.CurrentAnimation == NO_ANIMATION {
		return nil
	}
	animation := sa.Animation[sa.CurrentAnimation]
	ptr := gl.MapNamedBufferRange(
		sa.JointBufferObject,
		0,
		int(unsafe.Sizeof(y3d.Mat4{}))*len(animation.Joints),
		gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_RANGE_BIT,
	)
	if ptr == nil {
		return errors.New("cannot map joint buffer")
	}
	jointP := unsafe.Slice((*y3d.Mat4)(ptr), sa.MaxJoints)
	for i, j := range animation.Joints {
		jointP[i] = j.Transform.World.Mul(j.BindPose)
	}
	gl.UnmapNamedBuffer(sa.JointBufferObject)
	return nil
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
	isAnimated bool,
) *Joint {
	j := &Joint{
		Id:         id,
		Parent:     parent,
		Name:       name,
		BindPose:   bindPose,
		IsAnimated: isAnimated,
		Transform:  NewTransform(),
	}
	j.Transform.Position = pos
	j.Transform.Rotation = rot
	j.Transform.Scale = scale
	j.Transform.Recalulate()

	j.Transform.Local = j.Transform.Local.Mul(undeformedMatrix)
	return j
}

func (j *Joint) Update(joints []*Joint) {
	if j.Parent == NIL_JOINT_PARENT {
		j.Transform.World = j.Transform.Local
	} else {
		j.Transform.UpdateWorld(joints[j.Parent].Transform)
	}
	for _, c := range j.Children {
		joints[c].Update(joints)
	}
}

func (a *Animation) Destory() {
	clear(a.Joints)
}

func (s *SkeletalAnimator) BindBuffer() {
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, JOINTS_SSBO_BINDING, s.JointBufferObject)
}
