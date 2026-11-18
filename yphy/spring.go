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
