package yecs

import "yam/y3d"

const (
	CAM_TYPE_PERSPECTIVE = iota
	CAM_TYPE_ORTHOGRAPHIC
)

type Camera struct {
	Pos                                 y3d.Vec3
	Up                                  y3d.Vec3
	LookAt                              y3d.Vec3
	View                                y3d.Mat4
	NeedCalculation                     bool
	Speed                               float32
	CamType                             int
	Right, Left, Top, Bottom, Near, Far float32
}

func (c *Camera) GetViewTransformation() y3d.Mat4 {
	if c.NeedCalculation {
		z := y3d.Sub(c.LookAt, c.Pos)
		z = y3d.Normalize(z)

		x := y3d.Cross(z, c.Up)
		x = y3d.Normalize(x)

		y := y3d.Cross(x, z)
		y = y3d.Normalize(y)

		c.View = y3d.Mat4{x.X, x.Y, x.Z, 0.0,
			y.X, y.Y, y.Z, 0.0,
			z.X, z.Y, z.Z, 0.0,
			-c.Pos.X, -c.Pos.Y, -c.Pos.Z, 1.0,
		}
		c.NeedCalculation = false
	}
	return c.View
}

func (c *Camera) GetProjectionTransformation() y3d.Mat4 {
	switch c.CamType {
	case CAM_TYPE_ORTHOGRAPHIC:
		return y3d.Ortho(c.Left, c.Right, c.Bottom, c.Top, c.Near, c.Far)
	case CAM_TYPE_PERSPECTIVE:
		return y3d.Frustum(c.Left, c.Right, c.Bottom, c.Top, c.Near, c.Far)
	default:
		return y3d.Mat4{}
	}
}

func (c *Camera) Move(dt float64) {
	dir := y3d.Sub(c.LookAt, c.Pos)
	dir = y3d.Normalize(dir)

	vel := y3d.Smul(dir, c.Speed)
	c.Pos = y3d.Add(c.Pos, y3d.Smul(vel, float32(dt)))
	c.NeedCalculation = true
}
