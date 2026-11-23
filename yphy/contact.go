package yphy

import (
	"sync"
	"yam/y3d"
	"yam/ywd"
)

type Contact struct {
	World                   *ywd.World
	ContactNormal           y3d.Vec3 // A <- B
	ContactPosition         y3d.Vec3
	CoefficentOfRestitution float32
	Penetration             float32
	A, B                    ywd.EntityId
}

func (p *Contact) SetContactNormal(a, b ywd.EntityId) {
	posA := p.World.GetComponent(a, ywd.PositionComponent).(y3d.Vec3)
	posb := p.World.GetComponent(b, ywd.PositionComponent).(y3d.Vec3)
	p.ContactNormal = y3d.Normalize(y3d.Sub(posA, posb))
}

func (p *Contact) resolve(dt float32) {
	p.resolveVelocity(dt)
	p.resolveInterpenetration(dt)
}

func (p *Contact) calculateSeperatingVelocity() float32 {
	avel := p.World.GetComponent(p.A, ywd.VelocityComponent).(y3d.Vec3)
	if p.B != ywd.NullEntity {
		bvel := p.World.GetComponent(p.B,
			ywd.VelocityComponent).(y3d.Vec3)
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
	totalInverseMass := p.World.GetComponent(p.A, ywd.InverseMassComponent).(float32)
	if p.B != ywd.NullEntity {
		totalInverseMass += p.World.GetComponent(p.B, ywd.InverseMassComponent).(float32)
	}
	if totalInverseMass <= 0 {
		return
	}
	impluse := dv / totalInverseMass
	implusePerInverseMass := y3d.Smul(p.ContactNormal, impluse)
	velocityA := p.World.GetComponent(p.A, ywd.VelocityComponent).(y3d.Vec3)
	inverseMassA := p.World.GetComponent(p.A, ywd.InverseMassComponent).(float32)

	velocityA = y3d.Add(velocityA, y3d.Smul(implusePerInverseMass, inverseMassA))
	p.World.SetComponent(p.A, ywd.VelocityComponent, velocityA)
	if p.B != ywd.NullEntity {
		velocity := p.World.GetComponent(p.B, ywd.VelocityComponent).(y3d.Vec3)
		inverseMass := p.World.GetComponent(p.B, ywd.InverseMassComponent).(float32)
		velocity = y3d.Add(velocity, y3d.Smul(implusePerInverseMass, -inverseMass))
		p.World.SetComponent(p.B, ywd.VelocityComponent, velocity)
	}
}

func (p *Contact) resolveInterpenetration(dt float32) {

}

type ContactQueue struct {
	contacts []Contact
	mu       sync.Mutex
}

func NewContactQueue() *ContactQueue {
	return &ContactQueue{
		contacts: make([]Contact, 0),
	}
}

func (cq *ContactQueue) PushBack(contact Contact) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	cq.contacts = append(cq.contacts, contact)
}

func (cq *ContactQueue) Front() (Contact, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	if len(cq.contacts) == 0 {
		return Contact{}, false
	}
	c := cq.contacts[0]
	cq.contacts = cq.contacts[1:]
	return c, true
}
