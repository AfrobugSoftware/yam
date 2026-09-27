package ycontroller

import (
	"fmt"
	"math"
	"yam/y3d"
)

const (
	twoPI = float32(math.Pi * 2)
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

func NewMovementController() *MovementController {
	return &MovementController{
		Orientation: y3d.IdenQuat(),
		Right:       y3d.UNIT_X,
		Up:          y3d.UNIT_Y,
		Dir:         y3d.UNIT_Z,
	}
}

func (m *MovementController) RecalAxes() {
	if m.RotX > twoPI {
		m.RotX -= twoPI
	} else if m.RotX < -twoPI {
		m.RotX += twoPI
	}
	if m.RotY > twoPI {
		m.RotY -= twoPI
	} else if m.RotY < -twoPI {
		m.RotY += twoPI
	}
	if m.RotZ > twoPI {
		m.RotZ -= twoPI
	} else if m.RotZ < -twoPI {
		m.RotZ += twoPI
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
	m.RotX = m.RotSpeedPitch * deltaTime
	m.RotY = m.RotSpeedYaw * deltaTime
	m.RotZ = m.RotSpeedRoll * deltaTime
	if m.RotZ > 0 {
		fmt.Println(m.RotZ)
	}
	m.RecalAxes()

	m.Velocity = y3d.Smul(m.Dir, (m.Thrust * deltaTime))
	m.Position = y3d.Add(m.Position, m.Velocity)
}
