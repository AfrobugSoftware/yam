package ycontroller

import (
	"yam/y3d"
)

type MovementController struct {
	RotX             float32
	RotY             float32
	RotZ             float32
	RotSpeedRoll     float32
	RotSpeedPitch    float32
	RotSpeedYaw      float32
	RotSpeedRollMax  float32
	RotSpeedPitchMax float32
	RotSpeedYawMax   float32
	Dir              y3d.Vec3
	Up               y3d.Vec3
	Right            y3d.Vec3
	Thrust           float32
	Velocity         y3d.Vec3
	Orientation      y3d.Quaternion
	Position         y3d.Vec3
}

func NewMovementController(pos y3d.Vec3, orient y3d.Quaternion) *MovementController {
	return &MovementController{
		Orientation: y3d.IdenQuat(),
		Right:       y3d.UNIT_X,
		Up:          y3d.UNIT_Y,
		Dir:         y3d.UNIT_Z,
	}
}

func (m *MovementController) RecalAxes() {
	if m.RotX > y3d.TwoPI {
		m.RotX -= y3d.TwoPI
	} else if m.RotX < -y3d.TwoPI {
		m.RotX += y3d.TwoPI
	}
	if m.RotY > y3d.TwoPI {
		m.RotY -= y3d.TwoPI
	} else if m.RotY < -y3d.TwoPI {
		m.RotY += y3d.TwoPI
	}
	if m.RotZ > y3d.TwoPI {
		m.RotZ -= y3d.TwoPI
	} else if m.RotZ < -y3d.TwoPI {
		m.RotZ += y3d.TwoPI
	}
	frame := y3d.FromEuler(float64(m.RotX), float64(m.RotY), float64(m.RotZ))
	m.Orientation = y3d.ProdQuaternion(m.Orientation, frame)
	mR := m.Orientation.RotMat()

	m.Right = y3d.Vec3{
		X: mR[0],
		Y: mR[1],
		Z: mR[2],
	}
	m.Up = y3d.Vec3{
		X: mR[3],
		Y: mR[4],
		Z: mR[5],
	}
	m.Dir = y3d.Vec3{
		X: mR[6],
		Y: mR[7],
		Z: mR[8],
	}
}

func (m *MovementController) Update(deltaTime float32) {
	m.RotX = (m.RotSpeedPitch * deltaTime)
	m.RotY = (m.RotSpeedYaw * deltaTime)
	m.RotZ = (m.RotSpeedRoll * deltaTime)
	m.RecalAxes()

	m.Velocity = y3d.Smul(m.Dir, (m.Thrust * deltaTime))
	m.Position = y3d.Add(m.Position, m.Velocity)
}

func (m *MovementController) GetDir() y3d.Vec3 {
	return m.Dir
}

func (m *MovementController) GetUp() y3d.Vec3 {
	return m.Up
}
func (m *MovementController) GetRight() y3d.Vec3 {
	return m.Right
}
func (m *MovementController) GetPosition() y3d.Vec3 {
	return m.Position
}
func (m *MovementController) GetOrientation() y3d.Quaternion {
	return m.Orientation
}
