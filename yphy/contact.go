package yphy

import (
	"sort"
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
	p.resolveInterpenetration()
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

func (p *Contact) resolveVelocity(dt float32) {
	sv := p.calculateSeperatingVelocity()
	if sv > 0 {
		//stationary or seperating
		return
	}
	nsv := -sv * p.CoefficentOfRestitution
	accCausedVelocity := p.World.GetComponent(p.A, ywd.AccelerationComponent).(y3d.Vec3)
	if p.B != ywd.NullEntity {
		accCausedVelocity = y3d.Sub(accCausedVelocity,
			p.World.GetComponent(p.B, ywd.AccelerationComponent).(y3d.Vec3))
	}
	accCausedSepVelocity := y3d.Dot(accCausedVelocity, p.ContactNormal) * dt
	if accCausedSepVelocity < 0 {
		//we have closing velocity due to acceleration build up
		nsv += p.CoefficentOfRestitution * accCausedSepVelocity
		if nsv < 0 {
			nsv = 0
		}
	}
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

func (p *Contact) resolveInterpenetration() {
	if p.Penetration <= 0 {
		return
	}
	totalInverseMass := p.World.GetComponent(p.A, ywd.InverseMassComponent).(float32)
	if p.B != ywd.NullEntity {
		totalInverseMass += p.World.GetComponent(p.B, ywd.InverseMassComponent).(float32)
	}
	if totalInverseMass <= 0 {
		return //both particles have infinate mass
	}
	movePerIMass := y3d.Smul(p.ContactNormal, (p.Penetration / totalInverseMass))
	movement := y3d.Smul(movePerIMass, p.World.GetComponent(p.A, ywd.InverseMassComponent).(float32))
	posA := y3d.Add(p.World.GetComponent(p.A, ywd.PositionComponent).(y3d.Vec3), movement)
	p.World.SetComponent(p.A, ywd.PositionComponent, posA)
	if p.B != ywd.NullEntity {
		movement := y3d.Smul(movePerIMass, -p.World.GetComponent(p.B, ywd.InverseMassComponent).(float32))
		posb := y3d.Add(p.World.GetComponent(p.B, ywd.PositionComponent).(y3d.Vec3), movement)
		p.World.SetComponent(p.B, ywd.PositionComponent, posb)
	}
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

func (cq *ContactQueue) sort() {
	sort.Slice(cq.contacts, func(i, j int) bool {
		a := cq.contacts[i]
		b := cq.contacts[j]
		//most negative first
		return a.calculateSeperatingVelocity() <
			b.calculateSeperatingVelocity()
	})
}

func (cq *ContactQueue) ResolveContacts(dt float32) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	iterations := len(cq.contacts) * 2
	for range iterations {
		cq.sort()
		contact := cq.contacts[0]
		if contact.calculateSeperatingVelocity() > 0 {
			return //if the first contact is seperating we are done
		}
		contact.resolve(dt)
	}
}
