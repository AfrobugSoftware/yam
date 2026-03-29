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
	Planes                              []y3d.Plane
}

func (c *Camera) GetViewTransformation() y3d.Mat4 {
	if c.NeedCalculation {
		z := y3d.Sub(c.LookAt, c.Pos)
		z = y3d.Normalize(z)

		x := y3d.Cross(c.Up, z)
		x = y3d.Normalize(x)

		y := y3d.Cross(z, x)
		y = y3d.Normalize(y)
		//[Z,U,X]
		c.View = y3d.Mat4{x.X, y.X, z.X, 0.0,
			x.Y, y.Y, z.Y, 0.0,
			x.Z, y.Z, z.Z, 0.0,
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

func (c *Camera) ConstructPlanes() {
	//view := c.GetViewTransformation()

}

type CullSystem struct{}

func (c *CullSystem) Init()                                            {}
func (c *CullSystem) Shutdown()                                        {}
func (c *CullSystem) Update(w *World, dt float64, entities []EntityId) {}
func (c *CullSystem) Query() []ComponentId {
	return []ComponentId{AABBComponent}
}
