package yphy

import (
	"math"
	"yam/y3d"
)

type Particle struct {
	Position     y3d.Vec3
	Velocity     y3d.Vec3
	Acceleration y3d.Vec3
	InverseMass  float32
	Damping      float32
}

func (p *Particle) Update(dt float32) {
	p.Position = p.Position.AddScale(p.Velocity, dt)
	//update acceleration
	retAcc := p.Acceleration
	p.Velocity = p.Velocity.AddScale(retAcc, dt)

	p.Velocity = y3d.Smul(p.Velocity, float32(math.Pow(float64(p.Damping), float64(dt))))
}
