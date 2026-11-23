package yphy

import (
	"yam/y3d"
	"yam/ywd"
)

type Contact struct {
	World                   *ywd.World
	ContactNormal           y3d.Vec3 // A <- B
	ContactPosition         y3d.Vec3
	CoefficentOfRestitution float32
	InterpenerationDepth    float32
	A, B                    ywd.EntityId
}

func (p *Contact) SetContactNormal(a, b *Particle) {
	p.ContactNormal = y3d.Normalize(y3d.Sub(a.Position, b.Position))
}

func (p *Contact) resolve(dt float32) {
	p.resolveVelocity(dt)
	p.resolveInterpenetration(dt)
}

func (p *Contact) calculateSeperatingVelocity() float32 {
	avel := p.World.GetComponent(p.A, ywd.ParticleComponent).(Particle).Velocity
	if p.B != ywd.NullEntity {
		bvel := p.World.GetComponent(p.B,
			ywd.ParticleComponent).(Particle).Velocity
		avel = y3d.Sub(avel, bvel)
	}
	return y3d.Dot(avel, p.ContactNormal)
}

func (p *Contact) resolveVelocity(_ float32) {
	sv := p.calculateSeperatingVelocity()
	if sv > 0 {
		//stationary or seperating
		return
	}
	nsv := -sv * p.CoefficentOfRestitution
	dv := nsv - sv
	totalInverseMass := p.World.GetComponent(p.A, ywd.ParticleComponent).(Particle).InverseMass
	if p.B != ywd.NullEntity {
		totalInverseMass += p.World.GetComponent(p.B, ywd.ParticleComponent).(Particle).InverseMass
	}
	if totalInverseMass <= 0 {
		return
	}
	impluse := dv / totalInverseMass
	implusePerInverseMass := y3d.Smul(p.ContactNormal, impluse)
	particleA := p.World.GetComponent(p.A, ywd.ParticleComponent).(Particle)
	particleA.Velocity = y3d.Add(particleA.Velocity, y3d.Smul(implusePerInverseMass,
		particleA.InverseMass))
	p.World.SetComponent(p.A, ywd.ParticleComponent, particleA)
	if p.B != ywd.NullEntity {
		particleB := p.World.GetComponent(p.B, ywd.ParticleComponent).(Particle)
		particleB.Velocity = y3d.Add(particleB.Velocity, y3d.Smul(implusePerInverseMass,
			-particleB.InverseMass))
		p.World.SetComponent(p.B, ywd.ParticleComponent, particleB)
	}
}

func (p *Contact) resolveInterpenetration(dt float32) {

}
