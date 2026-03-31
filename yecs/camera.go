package yecs

import (
	"math"
	"yam/y3d"
)

const (
	CAM_TYPE_PERSPECTIVE = iota
	CAM_TYPE_ORTHOGRAPHIC
)

type Camera struct {
	Pos                                 y3d.Vec3
	Up                                  y3d.Vec3
	LookAt                              y3d.Vec3
	View                                y3d.Mat4
	Proj                                y3d.Mat4
	Speed                               float32
	CamType                             int
	Right, Left, Top, Bottom, Near, Far float32
	PitchSpeed, MaxPitch, Pitch         float32 //pitch is in degrees
	Planes                              []y3d.Plane
}

func (c *Camera) Recalulate() {
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
	switch c.CamType {
	case CAM_TYPE_ORTHOGRAPHIC:
		c.Proj = y3d.Ortho(c.Left, c.Right, c.Bottom, c.Top, c.Near, c.Far)
	case CAM_TYPE_PERSPECTIVE:
		c.Proj = y3d.Frustum(c.Left, c.Right, c.Bottom, c.Top, c.Near, c.Far)
	default:
		c.Proj = y3d.Identity
	}
}

func (c *Camera) PitchCamera() {
	//how to get the camera right
	right := y3d.Vec3{
		X: c.View[0], Y: c.View[4], Z: c.View[8],
	}
	rot := y3d.FromAngleAxis(right, float64(c.Pitch))
	c.LookAt = rot.RotateVec3(c.LookAt)
	c.Up = rot.RotateVec3(c.Up)

	c.Recalulate()
}

func (c *Camera) GetViewTransformation() y3d.Mat4 {
	return c.View
}

func (c *Camera) GetProjectionTransformation() y3d.Mat4 {
	return c.Proj
}

func (c *Camera) Move(dt float64) {
	dir := y3d.Sub(c.LookAt, c.Pos)
	dir = y3d.Normalize(dir)

	vel := y3d.Smul(dir, c.Speed)
	c.Pos = y3d.Add(c.Pos, y3d.Smul(vel, float32(dt)))

	c.Pitch += c.PitchSpeed * float32(dt)
	c.Pitch = y3d.Clamp(c.Pitch, -c.MaxPitch, c.MaxPitch)
	if math.Abs(float64(c.Pitch)) > y3d.NearZero {
		c.PitchCamera()
	}
	c.Recalulate()
}

func (c *Camera) ConstructPlanes() {
	//view := c.View

}

type CullSystem struct{}

func (c *CullSystem) Init()                                            {}
func (c *CullSystem) Shutdown()                                        {}
func (c *CullSystem) Update(w *World, dt float64, entities []EntityId) {}
func (c *CullSystem) Query() []ComponentId {
	return []ComponentId{AABBComponent}
}
