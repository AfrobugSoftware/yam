package yphy

import (
	"yam/y3d"
	"yam/ywd"
)

type Rod struct {
	W      *ywd.World
	A, B   ywd.EntityId
	Length float32
}

func (r *Rod) GenerateContact(cq *ContactQueue) {
	posA := r.W.GetComponent(r.A, ywd.PositionComponent).(y3d.Vec3)
	posB := r.W.GetComponent(r.B, ywd.PositionComponent).(y3d.Vec3)
	length := y3d.Sub(posA, posB).Length()

	if r.Length == length {
		return
	}
	normal := y3d.Normalize(y3d.Sub(posB, posA)) //pulling inwords not out
	contact := Contact{
		A: r.A,
		B: r.B,
	}
	if length > r.Length {
		contact.ContactNormal = normal
		contact.Penetration = length - r.Length
	} else {
		contact.ContactNormal = y3d.Smul(contact.ContactNormal, -1)
		contact.Penetration = r.Length - length
	}
	cq.PushBack(contact)
}
