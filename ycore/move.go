package ycore

import (
	"yam/y3d"
	"yam/ywd"
)

type Move struct {
	RotSpeedRoll  float32
	RotSpeedPitch float32
	RotSpeedYaw   float32
	ForwardSpeed  float32
	RightSpeed    float32
	UpSpeed       float32
}

func (m *Move) Update(w *ywd.World, e ywd.EntityId, dt float32) {
	position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
	orientation := w.GetComponent(e, ywd.OrientationComponent).(y3d.Quaternion)
	RotX := (m.RotSpeedPitch * dt)
	RotY := (m.RotSpeedYaw * dt)
	RotZ := (m.RotSpeedRoll * dt)
	if RotX > y3d.TwoPI {
		RotX -= y3d.TwoPI
	} else if RotX < -y3d.TwoPI {
		RotX += y3d.TwoPI
	}
	if RotY > y3d.TwoPI {
		RotY -= y3d.TwoPI
	} else if RotY < -y3d.TwoPI {
		RotY += y3d.TwoPI
	}
	if RotZ > y3d.TwoPI {
		RotZ -= y3d.TwoPI
	} else if RotZ < -y3d.TwoPI {
		RotZ += y3d.TwoPI
	}
	frame := y3d.FromEuler(RotX, RotY, RotZ)
	orientation = y3d.ProdQuaternion(orientation, frame)
	dir := y3d.Smul(orientation.Rotate(y3d.NegateVec3(y3d.UNIT_Z)), (m.ForwardSpeed * dt))
	right := y3d.Smul(orientation.Rotate(y3d.UNIT_X), (m.RightSpeed * dt))
	up := y3d.Smul(orientation.Rotate(y3d.UNIT_Y), (m.UpSpeed * dt))

	position = y3d.Add(position, dir, up, right)

	w.SetComponent(e, ywd.PositionComponent, position)
	w.SetComponent(e, ywd.OrientationComponent, orientation)
}
