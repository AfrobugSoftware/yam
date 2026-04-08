package yecs

import (
	"math"
	"yam/y3d"
)

const (
	CAM_TYPE_PERSPECTIVE = iota
	CAM_TYPE_ORTHOGRAPHIC
)

type CameraMode int

const (
	CAMERA_WORLD CameraMode = iota
	CAMERA_FOLLOW
	CAMERA_ORBIT
	CAMERA_SPLINE
)

type ICamera interface {
	GetProjectionTransformation()
	GetViewTransformation()
}

type Camera struct {
	Pos                                                  y3d.Vec3
	Up                                                   y3d.Vec3
	LookAt                                               y3d.Vec3
	View                                                 y3d.Mat4
	Proj                                                 y3d.Mat4
	Speed                                                float32
	CamType                                              int
	CamMode                                              CameraMode
	Entity                                               EntityId
	Right, Left, Top, Bottom, Near, Far                  float32
	PitchSpeed, MaxPitch, Pitch                          float32 //pitch is in degrees
	VerticalDistance, HorizontalDistance, TargetDistance float32
	SpringConstant                                       float32
	ActualPos, Velocity                                  y3d.Vec3
	YawSpeed                                             float32 //in radians/seconds
	Offset                                               y3d.Vec3
	Path                                                 y3d.Spline
	Indx                                                 int
	T                                                    float32
	Paused                                               bool
	Planes                                               []y3d.Plane
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

func (c Camera) GetViewTransformation() y3d.Mat4 {
	return c.View
}

func (c Camera) GetProjectionTransformation() y3d.Mat4 {
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

func (c Camera) Unproject(screenPoint y3d.Vec3, screenWidth, screenHeight int) y3d.Vec3 {
	dcX := screenPoint.X / float32(screenWidth) * 0.5
	dcY := screenPoint.Y / float32(screenHeight) * 0.5

	unprojection := c.Proj.Mul(c.View)
	(&unprojection).Invert()
	pw := unprojection.MulVec4(y3d.Vec4{X: dcX, Y: dcY, Z: screenPoint.Z, W: 1})
	return y3d.Vec3{
		X: pw.X / pw.W,
		Y: pw.Y / pw.W,
		Z: pw.Z / pw.W,
	}
}

// ray is from the center of the screen
func (c Camera) GetScreenRay(screenWidth, screenHeight int) y3d.Ray {
	point := y3d.Vec3{}
	start := c.Unproject(point, screenWidth, screenHeight)
	point.Z = 0.9
	end := c.Unproject(point, screenWidth, screenHeight)
	direction := y3d.Sub(end, start)
	direction = y3d.Normalize(direction)
	return y3d.Ray{
		O: start,
		D: direction,
	}
}

func (c *Camera) CullView(b Box, projView y3d.Mat4) bool {
	w := b.World
	p0 := w.Min
	p1 := y3d.Vec3{X: w.Min.X, Y: w.Max.Y, Z: w.Min.Z}
	p2 := y3d.Vec3{X: w.Min.X, Y: w.Max.Y, Z: w.Max.Z}
	p3 := y3d.Vec3{X: w.Min.X, Y: w.Min.Y, Z: w.Max.Z}
	p4 := y3d.Vec3{X: w.Max.X, Y: w.Max.Y, Z: w.Min.Z}
	p5 := y3d.Vec3{X: w.Max.X, Y: w.Min.Y, Z: w.Min.Z}
	p6 := y3d.Vec3{X: w.Max.X, Y: w.Min.Y, Z: w.Max.Z}
	p7 := w.Max

	for _, p := range []y3d.Vec3{p0, p1, p2, p3, p4, p5, p6, p7} {
		p4d := y3d.Vec4{X: p.X, Y: p.Y, Z: p.Z, W: 1}
		clipSpace := projView.MulVec4(p4d)
		isInside := (clipSpace.X <= clipSpace.W &&
			clipSpace.X >= -clipSpace.W &&
			clipSpace.Y <= clipSpace.W &&
			clipSpace.Y >= -clipSpace.W &&
			clipSpace.Z <= clipSpace.W &&
			clipSpace.Z >= -clipSpace.W)
		if isInside {
			return false
		}
	}
	return true
}

func (cam *Camera) UpdateFollow(w *World, dt float64, e EntityId) {
	dampening := 2.0 * math.Sqrt(float64(cam.SpringConstant))
	pos, targetPos := cam.ComputeFollowPosition(w)
	diff := y3d.Sub(cam.ActualPos, pos)

	acel := y3d.Sub(y3d.Smul(diff, -cam.SpringConstant), y3d.Smul(cam.Velocity, float32(dampening)))
	cam.Velocity = y3d.Add(cam.Velocity, y3d.Smul(acel, float32(dt)))
	cam.ActualPos = y3d.Add(cam.ActualPos, y3d.Smul(cam.Velocity, float32(dt)))

	cam.LookAt = targetPos
	cam.Pos = cam.ActualPos
	cam.Up = y3d.UNIT_Y
	cam.Recalulate()
	w.SetComponent(e, CameraComponent, cam)
}

func (cam *Camera) UpdateOrbit(w *World, dt float64, e EntityId) {
	transform, ok := w.GetComponent(cam.Entity, TransformComponent).(Transform)
	if !ok {
		return
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

	cam.Recalulate()
	w.SetComponent(e, CameraComponent, cam)
}

func (cam *Camera) UpdateSpline(w *World, dt float64, e EntityId) {
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

	cam.Recalulate()
	w.SetComponent(e, CameraComponent, cam)
}

func (c *Camera) ComputeFollowPosition(w *World) (pos, targetPos y3d.Vec3) {
	transform := w.GetComponent(c.Entity, TransformComponent).(Transform)
	pos = transform.Position
	pos = y3d.Sub(c.Pos, y3d.Smul(transform.GetForward(), c.HorizontalDistance))
	pos = y3d.Add(c.Pos, y3d.Smul(transform.GetUp(), c.VerticalDistance))
	targetPos = y3d.Add(transform.Position, y3d.Smul(transform.GetForward(), c.TargetDistance))

	return
}

func (c *Camera) SnapToIdeal(w *World) {
	pos, target := c.ComputeFollowPosition(w)
	c.ActualPos = pos
	c.LookAt = target
	c.Velocity = y3d.ZEROV

	c.Recalulate()
}

type CameraSystem struct{}

func (c *CameraSystem) Init()     {}
func (c *CameraSystem) Shutdown() {}
func (c *CameraSystem) Update(w *World, dt float64, entites []EntityId) {
	for _, e := range entites {
		cam := w.GetComponent(e, CameraComponent).(Camera)
		c := &cam
		switch c.CamMode {
		case CAMERA_WORLD:
			continue
		case CAMERA_FOLLOW:
			c.UpdateFollow(w, dt, e)
		case CAMERA_ORBIT:
			c.UpdateOrbit(w, dt, e)
		case CAMERA_SPLINE:
			c.UpdateSpline(w, dt, e)
		}
	}
}
func (c *CameraSystem) Query() []ComponentId {
	return []ComponentId{CameraComponent}
}
