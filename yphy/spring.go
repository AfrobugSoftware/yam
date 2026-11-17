package yphy

import (
	"math"
	"yam/y3d"
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

func (as *AnchorSpringForce) Force(p *Particle, dt float32) y3d.Vec3 {
	d := y3d.Sub(p.Position, as.anchor)
	dp := y3d.Normalize(d)
	dist := d.Length()
	coeff := as.k * float32(math.Abs(float64(dist-as.restingLength)))
	return y3d.Smul(dp, -coeff)
}

type SpringForce struct {
	Other          *Particle
	SpringConstant float32
	RestLength     float32
}

func NewSpringForce(other *Particle, springConstant, restLength float32) *SpringForce {
	return &SpringForce{
		Other:          other,
		SpringConstant: springConstant,
		RestLength:     restLength,
	}
}

func (s *SpringForce) Force(p *Particle, dt float32) y3d.Vec3 {
	d := y3d.Sub(p.Position, s.Other.Position)
	dp := y3d.Normalize(d)
	dist := d.Length()
	coeff := s.SpringConstant * float32(math.Abs(float64(dist-s.RestLength)))
	return y3d.Smul(dp, -coeff)
}
