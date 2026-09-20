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
	StartFrame   int
	EndFrame     int
	CurFrame     int
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
	starframe, endframe int,
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

type Joint struct {
	Id         int
	Parent     int
	Name       string
	Children   []int
	kTime      []float32
	KPos       []y3d.Vec3
	KRot       []y3d.Quaternion
	Transform  *Transform
	IsAnimated bool
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
		CurrentAnimation: -1,
	}
	gl.CreateBuffers(1, &sa.JointBufferObject)
	gl.NamedBufferStorage(sa.JointBufferObject, int(unsafe.Sizeof(y3d.Mat4{}))*maxJoints, nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	return sa
}

func (sa *SkeletalAnimator) Add(animation *Animation) (int, error) {
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
	animation.CurFrame = 0
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
	animation.CurFrame = max(animation.StartFrame+
		int(animation.FPS*animation.PlaybackTime), animation.StartFrame)
	animation.PlaybackTime += deltaTime
	if animation.CurFrame >= animation.EndFrame {
		animation.IsComplete = true
		animation.CurFrame = animation.StartFrame
		animation.PlaybackTime = 0
	} else {
		animation.IsComplete = false
		if animation.CurFrame != animation.StartFrame {
			for _, j := range animation.Joints {
				//this is not correct
				pos := y3d.Lerp(j.KPos[animation.CurFrame],
					j.KPos[animation.CurFrame+1], deltaTime)
				rot := y3d.Slerp(j.KRot[animation.CurFrame],
					j.KRot[animation.CurFrame+1], float64(deltaTime))
				j.Transform.Position = pos
				j.Transform.Rotation = rot
				j.Transform.Recalulate()
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
		jointP[i] = j.Transform.World
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
) *Joint {
	j := &Joint{
		Id:        id,
		Parent:    parent,
		Name:      name,
		Transform: NewTransform(),
	}
	j.Transform.Position = pos
	j.Transform.Rotation = rot
	j.Transform.Recalulate()
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
