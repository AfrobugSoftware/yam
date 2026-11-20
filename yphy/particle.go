package yphy

import (
	"math"
	"yam/y3d"
)

type Particle struct {
	Position     y3d.Vec3
	Velocity     y3d.Vec3
	Acceleration y3d.Vec3
	ForceAccum   y3d.Vec3
	InverseMass  float32
	Damping      float32
}

func (p *Particle) HasInfinateMass() bool {
	return p.InverseMass == 0
}

func (p *Particle) Update(dt float32) {
	if p.InverseMass <= 0.0 {
		return
	}
	p.Position = p.Position.AddScale(p.Velocity, dt)
	//update acceleration
	retAcc := p.Acceleration
	retAcc = retAcc.AddScale(p.ForceAccum, p.InverseMass)
	p.Velocity = p.Velocity.AddScale(retAcc, dt)

	p.Velocity = y3d.Smul(p.Velocity, float32(math.Pow(float64(p.Damping), float64(dt))))
	p.ForceAccum = y3d.UnitV3
}

func (p *Particle) AddForce(f ForceGenerator, dt float32) {
	p.ForceAccum = y3d.Add(p.ForceAccum, f.Force(p, dt))
}
