package ygame

import "yam/y3d"

type Camera struct {
	Pos                                 y3d.Vec3
	Up                                  y3d.Vec3
	LookAt                              y3d.Vec3
	View                                y3d.Mat4
	NeedCalculation                     bool
	Speed                               float64
	Right, Left, Top, Bottom, Near, Far float64
}

func (c *Camera) GetViewTransformation() y3d.Mat4 {
	if c.NeedCalculation {
		z := y3d.Sub(c.LookAt, c.Pos)
		z = y3d.Normalize(z)

		x := y3d.Cross(z, c.Up)
		x = y3d.Normalize(x)

		y := y3d.Cross(x, z)
		y = y3d.Normalize(y)

		c.View = y3d.Mat4{x.X, y.X, z.X, 0.0,
			x.Y, y.Y, z.Y, 0.0,
			x.Z, y.Z, z.Z, 0.0,
			-c.Pos.X, -c.Pos.Y, -c.Pos.Z, 1.0,
		}
		c.NeedCalculation = false
	}
	return c.View
}

func (c *Camera) GetProjTransformation() y3d.Mat4 {
	return y3d.Frustum(c.Left, c.Right, c.Bottom, c.Top, c.Near, c.Far)
}

func (c *Camera) Move(dt float64) {
	dir := y3d.Sub(c.LookAt, c.Pos)
	dir = y3d.Normalize(dir)

	vel := y3d.Smul(dir, c.Speed)
	c.Pos = y3d.Add(c.Pos, y3d.Smul(vel, dt))
	c.NeedCalculation = true
}
