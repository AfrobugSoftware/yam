package ycore

import "yam/y3d"

const (
	NIL_JOINT_PARENT = -1
	ROOT             = 0 //root is alway at zero
)

type Joint struct {
	Id         int
	Parent     int
	Children   []int
	Pos        y3d.Vec3
	Rot        y3d.Quaternion
	kTime      []float32
	KPos       []y3d.Vec3
	KRot       []y3d.Quaternion
	Matrix     y3d.Mat3
	Bind       y3d.Mat3
	IsAnimated bool
}

type SkeletalAnimator struct {
	Joints []*Joint
}
