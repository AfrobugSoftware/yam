package yphy

import (
	"yam/y3d"
	"yam/ywd"
)

type Cable struct {
	W          *ywd.World
	A, B       ywd.EntityId
	MaxLength  float32
	Resitution float32
}

func NewCable(w *ywd.World, a, b ywd.EntityId, maxLength, resitution float32) *Cable {
	return &Cable{
		W:          w,
		A:          a,
		B:          b,
		MaxLength:  maxLength,
		Resitution: resitution,
	}
}

func (c *Cable) GenerateContact(cq *ContactQueue) {
	posA := c.W.GetComponent(c.A, ywd.PositionComponent).(y3d.Vec3)
	posB := c.W.GetComponent(c.B, ywd.PositionComponent).(y3d.Vec3)
	length := y3d.Sub(posA, posB).Length()

	if length < c.MaxLength {
		return
	}
	contact := Contact{
		A:                       c.A,
		B:                       c.B,
		CoefficentOfRestitution: c.Resitution,
		Penetration:             length - c.MaxLength,
	}
	normal := y3d.Normalize(y3d.Sub(posB, posA)) //pulling inwords not out
	contact.ContactNormal = normal
	cq.PushBack(contact)
}
