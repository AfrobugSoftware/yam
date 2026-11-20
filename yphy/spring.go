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
func (as *AnchorSpringForce) Force(p *Particle, dt float32) y3d.Vec3 {
	d := y3d.Sub(p.Position, as.anchor)
	dp := y3d.Normalize(d)
	dist := d.Length()
	coeff := as.k * (as.restingLength - dist)
	return y3d.Smul(dp, coeff)
}

type SpringForce struct {
	World          *ywd.World
	Other          ywd.EntityId
	SpringConstant float32
	RestLength     float32
}

func NewSpringForce(world *ywd.World, other ywd.EntityId, springConstant, restLength float32) *SpringForce {
	return &SpringForce{
		World:          world,
		Other:          other,
		SpringConstant: springConstant,
		RestLength:     restLength,
	}
}

func (s *SpringForce) Force(p *Particle, dt float32) y3d.Vec3 {
	otherPos := s.World.GetComponent(s.Other, ywd.ParticleComponent).(Particle).Position
	d := y3d.Sub(p.Position, otherPos)
	dp := y3d.Normalize(d)
	dist := d.Length()
	coeff := s.SpringConstant * float32(math.Abs(float64(dist-s.RestLength)))
	return y3d.Smul(dp, -coeff)
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

func (s *BungeeSpringForce) Force(p *Particle, dt float32) y3d.Vec3 {
	otherPos := s.World.GetComponent(s.Other, ywd.ParticleComponent).(Particle).Position
	d := y3d.Sub(p.Position, otherPos)
	dist := d.Length()
	if dist <= s.RestLength {
		return y3d.UnitV3
	}
	dp := y3d.Normalize(d)
	coeff := s.SpringConstant * (s.RestLength - dist)
	return y3d.Smul(dp, -coeff)
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
func (s *FakeSpringForce) Force(p *Particle, dt float32) y3d.Vec3 {
	if p.InverseMass == 0 {
		return y3d.Vec3{} //infinte mass
	}

	position := y3d.Sub(p.Position, s.Anchor)
	gamma := float32(0.5 * math.Sqrt(4*float64(s.SpringConstant)-float64(s.Damping*s.Damping)))
	if gamma == 0 {
		return y3d.UnitV3
	}
	c := y3d.Add(y3d.Smul(position, (s.Damping/(2*float32(gamma)))), y3d.Smul(p.Velocity, 1/float32(gamma)))
	target := y3d.Add(y3d.Smul(position, float32(math.Cos(float64(gamma*dt)))), y3d.Smul(c, float32(math.Sin(float64(gamma*dt)))))
	target = y3d.Smul(target, float32(math.Exp(float64(-0.5*dt*s.Damping))))
	accel := y3d.Smul(y3d.Sub(target, position), 1/(dt*dt))
	accel = y3d.Sub(accel, y3d.Smul(p.Velocity, dt))
	return y3d.Smul(accel, 1/p.InverseMass)
}
