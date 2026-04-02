package yecs

import (
	"math"
	"yam/y3d"
)

const (
	CAM_TYPE_PERSPECTIVE = iota
	CAM_TYPE_ORTHOGRAPHIC
)

type ICamera interface {
	GetProjectionTransformation()
	GetViewTransformation()
}

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

func (c *Camera) Update(dt float64) {
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

func (c Camera) Unproject(x, y float32, screenWidth, screenHeight int) y3d.Vec3 {
	dcX := x / float32(screenWidth) * 0.5
	dcY := y / float32(screenHeight) * 0.5

	unprojection := c.Proj.Mul(c.View)
	(&unprojection).Invert()
	pw := unprojection.MulVec4(y3d.Vec4{X: dcX, Y: dcY, Z: 0, W: 1})
	return y3d.Vec3{
		X: pw.X / pw.W,
		Y: pw.Y / pw.W,
		Z: pw.Z / pw.W,
	}
}

func (c *Camera) ConstructPlanes() {
	//view := c.View

}

type FollowCamera struct {
	Camera
	EntityToFollow                                       EntityId
	VerticalDistance, HorizontalDistance, TargetDistance float32
	SpringConstant                                       float32
	ActualPos, Velocity                                  y3d.Vec3
}

func (c *FollowCamera) ComputeFollowPosition(w *World) (pos, targetPos y3d.Vec3) {
	transform := w.GetComponent(c.EntityToFollow, TransformComponent).(Transform)
	pos = transform.Position
	pos = y3d.Sub(c.Pos, y3d.Smul(transform.GetForward(), c.HorizontalDistance))
	pos = y3d.Add(c.Pos, y3d.Smul(transform.GetUp(), c.VerticalDistance))
	targetPos = y3d.Add(transform.Position, y3d.Smul(transform.GetForward(), c.TargetDistance))

	return
}

func (c *FollowCamera) SnapToIdel(w *World) {
	pos, target := c.ComputeFollowPosition(w)
	c.ActualPos = pos
	c.LookAt = target
	c.Velocity = y3d.ZEROV

	c.Recalulate()
}

type FollowCameraSystem struct{}

func (c *FollowCameraSystem) Init()     {}
func (c *FollowCameraSystem) Shutdown() {}
func (c *FollowCameraSystem) Update(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		cam := w.GetComponent(e, FollowCameraComponent).(FollowCamera)
		dampening := 2.0 * math.Sqrt(float64(cam.SpringConstant))
		pos, targetPos := cam.ComputeFollowPosition(w)
		diff := y3d.Sub(cam.ActualPos, pos)

		acel := y3d.Sub(y3d.Smul(diff, -cam.SpringConstant), y3d.Smul(cam.Velocity, float32(dampening)))
		cam.Velocity = y3d.Add(cam.Velocity, y3d.Smul(acel, float32(dt)))
		cam.ActualPos = y3d.Add(cam.ActualPos, y3d.Smul(cam.Velocity, float32(dt)))

		cam.LookAt = targetPos
		cam.Pos = cam.ActualPos
		cam.Up = y3d.UNIT_Y
		(&cam).Recalulate()

		w.SetComponent(e, FollowCameraComponent, cam)
	}
}
func (c *FollowCameraSystem) Query() []ComponentId {
	return []ComponentId{FollowCameraComponent}
}

type OrbitCamera struct {
	Camera
	EntityToOrbit        EntityId
	PitchSpeed, YawSpeed float32 //in radians/seconds
	Offset               y3d.Vec3
}

type OrbitCameraSystem struct{}

func (or *OrbitCameraSystem) Init()     {}
func (or *OrbitCameraSystem) Shutdown() {}
func (or *OrbitCameraSystem) Query() []ComponentId {
	return []ComponentId{OrbitCameraComponent}
}
func (or *OrbitCameraSystem) Update(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		cam := w.GetComponent(e, OrbitCameraComponent).(OrbitCamera)
		transform, ok := w.GetComponent(cam.EntityToOrbit, TransformComponent).(Transform)
		if !ok {
			continue
		}
		yaw := y3d.FromAngleAxis(y3d.UNIT_Y, float64(cam.YawSpeed)*dt)

		cam.Offset = yaw.RotateVec3(cam.Offset)
		cam.Up = yaw.RotateVec3(cam.Up)

		forward := y3d.NegateVec3(cam.Offset)
		forward = y3d.Normalize(forward)
		right := y3d.Cross(cam.Up, forward)

		pitch := y3d.FromAngleAxis(right, float64(cam.PitchSpeed)*dt)
		cam.Offset = pitch.RotateVec3(cam.Offset)
		cam.Up = pitch.RotateVec3(cam.Up)
		cam.LookAt = transform.Position

		(&cam).Recalulate()
		w.SetComponent(e, OrbitCameraComponent, cam)
	}
}

type SplineCamera struct {
	Camera
	Path   y3d.Spline
	Indx   int
	T      float32
	Paused bool
}
type SplineCameraSystem struct{}

func (sc *SplineCameraSystem) Init()     {}
func (sc *SplineCameraSystem) Shutdown() {}
func (sc *SplineCameraSystem) Query() []ComponentId {
	return []ComponentId{SplineCamaraComponent}
}
func (sc *SplineCameraSystem) Update(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		cam := w.GetComponent(e, SplineCamaraComponent).(SplineCamera)
		if !cam.Paused {
			cam.T = cam.Speed * float32(dt)
			if cam.T >= 1.0 {
				l := len(cam.Path.ControlPoints)
				if cam.Indx < l-3 {
					cam.Indx++
					cam.T = cam.T - 1.0
				} else {
					cam.Paused = true
				}
			}
		}
		cam.Pos = cam.Path.Compute(cam.Indx, cam.T)
		cam.LookAt = cam.Path.Compute(cam.Indx, cam.T+0.01)
		cam.Up = y3d.UNIT_Y

		(&cam).Recalulate()
		w.SetComponent(e, SplineCamaraComponent, cam)
	}
}
