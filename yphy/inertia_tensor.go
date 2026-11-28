package yphy

import "yam/y3d"

type InertiaTensor struct {
	Local y3d.Mat3
	World y3d.Mat3
}

func NewInertiaTensor(tensor y3d.Mat3) InertiaTensor {
	return InertiaTensor{
		Local: tensor.Inverse(),
		World: y3d.Identity3x3,
	}
}

//this transform should only orient the tensor
//there is no scale or translation
func (it *InertiaTensor) Transform(q y3d.Quaternion) {
	m := q.RotMat()
	t4 := m[0]*it.Local[0] + m[1]*it.Local[3] + m[2]*it.Local[6]
	t9 := m[0]*it.Local[1] + m[1]*it.Local[4] + m[2]*it.Local[7]
	t14 := m[0]*it.Local[2] + m[1]*it.Local[5] + m[2]*it.Local[8]

	t28 := m[3]*it.Local[0] + m[4]*it.Local[3] + m[5]*it.Local[6]
	t33 := m[3]*it.Local[1] + m[4]*it.Local[4] + m[5]*it.Local[7]
	t38 := m[3]*it.Local[2] + m[4]*it.Local[5] + m[5]*it.Local[8]

	t52 := m[6]*it.Local[0] + m[7]*it.Local[3] + m[8]*it.Local[6]
	t57 := m[6]*it.Local[1] + m[7]*it.Local[4] + m[8]*it.Local[7]
	t62 := m[6]*it.Local[2] + m[7]*it.Local[5] + m[8]*it.Local[8]

	it.World[0] = t4*m[0] + t9*m[1] + t14*m[2]
	it.World[1] = t4*m[3] + t9*m[4] + t14*m[5]
	it.World[2] = t4*m[6] + t9*m[7] + t14*m[8]
	it.World[3] = t28*m[0] + t33*m[1] + t38*m[2]
	it.World[4] = t28*m[3] + t33*m[4] + t38*m[5]
	it.World[5] = t28*m[6] + t33*m[7] + t38*m[8]
	it.World[6] = t52*m[0] + t57*m[1] + t62*m[2]
	it.World[7] = t52*m[3] + t57*m[4] + t62*m[5]
	it.World[8] = t52*m[6] + t57*m[7] + t62*m[8]
}
