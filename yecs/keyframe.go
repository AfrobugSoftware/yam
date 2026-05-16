package yecs

import "yam/y3d"

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

// t  = [0, 1]
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
