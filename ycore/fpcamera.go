package ycore

import (
	"yam/y3d"
	"yam/ycontroller"
)

type FPCamera struct {
	ycontroller.MovementController
	JumpSpeed    float32
	SlideSpeed   float32
	ForwardSpeed float32
}

func NewCamera(pos y3d.Vec3) *FPCamera {
	return &FPCamera{
		MovementController: *ycontroller.NewMovementController(
			pos, y3d.IdenQuat(),
		),
	}
}

func (c *FPCamera) RecalAxis() {
	if c.RotY > y3d.TwoPI {
		c.RotY -= y3d.TwoPI
	} else if c.RotY < -y3d.TwoPI {
		c.RotY += y3d.TwoPI
	}

	if c.RotX > 1.4 {
		c.RotX -= 1.4
	} else if c.RotX < -1.4 {
		c.RotX += 1.4
	}
	c.Right = y3d.UNIT_X
	c.Up = y3d.UNIT_Y
	c.Dir = y3d.UNIT_Z

	m := y3d.RotationAxis(c.Up, float64(c.RotY))
	c.Right = m.MulVec3(c.Right)
	c.Dir = m.MulVec3(c.Dir)

	m = y3d.RotationAxis(c.Right, float64(c.RotX))
	c.Up = m.MulVec3(c.Up)
	c.Dir = m.MulVec3(c.Dir)

	c.Dir = y3d.Normalize(c.Dir)
	c.Right = y3d.Normalize(y3d.Cross(c.Up, c.Dir))
	c.Up = y3d.Normalize(y3d.Cross(c.Dir, c.Right))
}

func (c *FPCamera) Update(deltaTime float32) {
	c.RotX += (c.RotSpeedPitch * deltaTime)
	c.RotY += (c.RotSpeedYaw * deltaTime)
	c.RotZ += (c.RotSpeedRoll * deltaTime)
	c.RecalAxis()

	v := y3d.Smul(c.Dir, (c.ForwardSpeed * deltaTime))
	u := y3d.Smul(c.Up, (c.JumpSpeed * deltaTime))
	s := y3d.Smul(c.Right, (c.SlideSpeed * deltaTime))

	c.Position = y3d.Add(c.Position, v, u, s)
}
