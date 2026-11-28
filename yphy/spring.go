package yphy

import (
	"math"
	"yam/y3d"
	"yam/ywd"
)

type AnchorSpringForce struct {
	anchor        y3d.Vec3
	k             float32
	restingLength float32
}

func NewAnchorSpringForce(point y3d.Vec3, k float32, restingLength float32) *AnchorSpringForce {
	return &AnchorSpringForce{
		anchor:        point,
		k:             k,
		restingLength: restingLength,
	}
}
func (as *AnchorSpringForce) SetAnchor(point y3d.Vec3) {
	as.anchor = point
}
func (as *AnchorSpringForce) Force(w *ywd.World, e ywd.EntityId, dt float32) {
	position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
	force := w.GetComponent(e, ywd.ForceComponent).(y3d.Vec3)
	d := y3d.Sub(position, as.anchor)
	dp := y3d.Normalize(d)
	dist := d.Length()
	coeff := as.k * (as.restingLength - dist)
	f := y3d.Smul(dp, coeff)
	force = y3d.Add(force, f)
	w.SetComponent(e, ywd.ForceComponent, force)
}

type SpringForce struct {
	World             *ywd.World
	Other             ywd.EntityId
	ConnectPoint      y3d.Vec3
	OtherConnectPoint y3d.Vec3
	SpringConstant    float32
	RestLength        float32
}

func NewSpringForce(world *ywd.World, other ywd.EntityId, springConstant, restLength float32) *SpringForce {
	return &SpringForce{
		World:          world,
		Other:          other,
		SpringConstant: springConstant,
		RestLength:     restLength,
	}
}

func (s *SpringForce) Force(w *ywd.World, e ywd.EntityId, dt float32) {
	position := TransformPointToWorldSpace(w, e, s.ConnectPoint)
	otherPos := TransformPointToWorldSpace(w, s.Other, s.OtherConnectPoint)
	force := y3d.Sub(position, otherPos)
	dist := force.Length()
	coeff := s.SpringConstant * float32(math.Abs(float64(dist-s.RestLength)))
	force = y3d.Normalize(force)
	f := y3d.Smul(force, -coeff)
	AddForceAtPoint(w, e, f, position)
}

type BungeeSpringForce struct {
	World          *ywd.World
	Other          ywd.EntityId
	SpringConstant float32
	RestLength     float32
}

func NewBungeeSpringForce(world *ywd.World, other ywd.EntityId, springConstant, restLength float32) *BungeeSpringForce {
	return &BungeeSpringForce{
		World:          world,
		Other:          other,
		SpringConstant: springConstant,
		RestLength:     restLength,
	}
}

func (s *BungeeSpringForce) Force(w *ywd.World, e ywd.EntityId, dt float32) {
	position := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
	otherPos := s.World.GetComponent(s.Other, ywd.PositionComponent).(y3d.Vec3)
	force := w.GetComponent(e, ywd.ForceComponent).(y3d.Vec3)
	d := y3d.Sub(position, otherPos)
	dist := d.Length()
	f := y3d.Vec3{}
	if dist <= s.RestLength {
		f = y3d.UnitV3
	} else {
		dp := y3d.Normalize(d)
		coeff := s.SpringConstant * (s.RestLength - dist)
		f = y3d.Smul(dp, -coeff)
	}
	force = y3d.Add(force, f)
	w.SetComponent(e, ywd.ForceComponent, force)
}

type FakeSpringForce struct {
	Anchor         y3d.Vec3
	SpringConstant float32
	Damping        float32
}

func NewFakeSpringForce(anchor y3d.Vec3, springConstant, damping float32) *FakeSpringForce {
	return &FakeSpringForce{
		Anchor:         anchor,
		SpringConstant: springConstant,
		Damping:        damping,
	}
}

// using damped harmonic motion to calculate the force of a spring with damping
// cannot be used for springs that has other forces acting on it, like gravity or other springs
func (s *FakeSpringForce) Force(w *ywd.World, e ywd.EntityId, dt float32) {
	pos := w.GetComponent(e, ywd.PositionComponent).(y3d.Vec3)
	velocity := w.GetComponent(e, ywd.VelocityComponent).(y3d.Vec3)
	inverseMass := w.GetComponent(e, ywd.InverseMassComponent).(float32)
	force := w.GetComponent(e, ywd.ForceComponent).(y3d.Vec3)
	if inverseMass == 0 {
		return
	}

	position := y3d.Sub(pos, s.Anchor)
	gamma := float32(0.5 * math.Sqrt(4*float64(s.SpringConstant)-float64(s.Damping*s.Damping)))
	f := y3d.Vec3{}
	if gamma == 0 {
		f = y3d.UnitV3
	} else {
		c := y3d.Add(y3d.Smul(position, (s.Damping/(2*float32(gamma)))), y3d.Smul(velocity, 1/float32(gamma)))
		target := y3d.Add(y3d.Smul(position, float32(math.Cos(float64(gamma*dt)))), y3d.Smul(c, float32(math.Sin(float64(gamma*dt)))))
		target = y3d.Smul(target, float32(math.Exp(float64(-0.5*dt*s.Damping))))
		accel := y3d.Smul(y3d.Sub(target, position), 1/(dt*dt))
		accel = y3d.Sub(accel, y3d.Smul(velocity, dt))
		f = y3d.Smul(accel, 1/inverseMass)
	}
	force = y3d.Add(force, f)
	w.SetComponent(e, ywd.ForceComponent, force)
}
