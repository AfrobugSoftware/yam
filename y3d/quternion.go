package y3d

import (
	"math"
)

// New returns a new quaternion
func NewQuaternion(w, x, y, z float64) Quaternion {
	return Quaternion{W: w, X: x, Y: y, Z: z}
}

func IdenQuat() Quaternion {
	return Quaternion{W: 1}
}

func Pure(x, y, z float64) Quaternion {
	return Quaternion{X: x, Y: y, Z: z}
}

// A quternion
type Quaternion struct {
	W, X, Y, Z float64
}

func (qin Quaternion) Conj() Quaternion {
	qin.X = -qin.X
	qin.Y = -qin.Y
	qin.Z = -qin.Z
	return qin
}

func (qin Quaternion) Norm2() float64 {
	return qin.W*qin.W + qin.X*qin.X + qin.Y*qin.Y + qin.Z*qin.Z
}

func (qin Quaternion) Neg() Quaternion {
	qin.W = -qin.W
	qin.X = -qin.X
	qin.Y = -qin.Y
	qin.Z = -qin.Z
	return qin
}

func (qin Quaternion) Norm() float64 {
	return math.Sqrt(qin.Norm2())
}

func Scalar(w float64) Quaternion {
	return Quaternion{W: w}
}

func SumQuaternion(qin ...Quaternion) Quaternion {
	qout := Quaternion{}
	for _, q := range qin {
		qout.W += q.W
		qout.X += q.X
		qout.Y += q.Y
		qout.Z += q.Z
	}
	return qout
}

func FromAngleAxis(axis Vec3, angle float64) Quaternion {
	axis = Normalize(axis)
	s := math.Sin(angle / 2)
	w := math.Cos(angle / 2)
	x := float64(axis.X) * s
	y := float64(axis.Y) * s
	z := float64(axis.Z) * s
	return NewQuaternion(w, x, y, z)
}

func ProdQuaternion(qin ...Quaternion) Quaternion {
	qout := Quaternion{1, 0, 0, 0}
	var w, x, y, z float64
	for _, q := range qin {
		w = qout.W*q.W - qout.X*q.X - qout.Y*q.Y - qout.Z*q.Z
		x = qout.W*q.X + qout.X*q.W + qout.Y*q.Z - qout.Z*q.Y
		y = qout.W*q.Y + qout.Y*q.W + qout.Z*q.X - qout.X*q.Z
		z = qout.W*q.Z + qout.Z*q.W + qout.X*q.Y - qout.Y*q.X
		qout = Quaternion{w, x, y, z}
	}
	return qout
}

func (qin Quaternion) Unit() Quaternion {
	k := qin.Norm()
	return Quaternion{qin.W / k, qin.X / k, qin.Y / k, qin.Z / k}
}

func (qin Quaternion) Inv() Quaternion {
	k2 := qin.Norm2()
	q := qin.Conj()
	return Quaternion{q.W / k2, q.X / k2, q.Y / k2, q.Z / k2}
}

func (qin Quaternion) RotateVec3(vec Vec3) Vec3 {
	conj := qin.Conj()
	aug := Quaternion{0,
		float64(vec.X),
		float64(vec.Y),
		float64(vec.Z),
	}
	rot := ProdQuaternion(qin, aug, conj)
	return Vec3{float32(rot.X), float32(rot.Y), float32(rot.Z)}
}

func Rotate(q Quaternion, vin Vec3) Vec3 {
	conj := q.Conj()
	aug := Quaternion{0,
		float64(vin.X),
		float64(vin.Y),
		float64(vin.Z),
	}
	rot := ProdQuaternion(q, aug, conj)
	return Vec3{
		float32(rot.X),
		float32(rot.Y),
		float32(rot.Z)}
}

func (q Quaternion) Euler() (float64, float64, float64) {
	r := q.Unit()
	phi := math.Atan2(2*(r.W*r.X+r.Y*r.Z), 1-2*(r.X*r.X+r.Y*r.Y))
	theta := math.Asin(2 * (r.W*r.Y - r.Z*r.X))
	psi := math.Atan2(2*(r.X*r.Y+r.W*r.Z), 1-2*(r.Y*r.Y+r.Z*r.Z))
	return phi, theta, psi
}

func FromEuler(phi, theta, psi float64) Quaternion {
	q := Quaternion{}
	q.W = math.Cos(phi/2)*math.Cos(theta/2)*math.Cos(psi/2) +
		math.Sin(phi/2)*math.Sin(theta/2)*math.Sin(psi/2)
	q.X = math.Sin(phi/2)*math.Cos(theta/2)*math.Cos(psi/2) -
		math.Cos(phi/2)*math.Sin(theta/2)*math.Sin(psi/2)
	q.Y = math.Cos(phi/2)*math.Sin(theta/2)*math.Cos(psi/2) +
		math.Sin(phi/2)*math.Cos(theta/2)*math.Sin(psi/2)
	q.Z = math.Cos(phi/2)*math.Cos(theta/2)*math.Sin(psi/2) -
		math.Sin(phi/2)*math.Sin(theta/2)*math.Cos(psi/2)
	return q
}

// role major
func (qin Quaternion) RotMat() Mat3 {
	q := qin.Unit()
	w, x, y, z := float32(q.W), float32(q.X), float32(q.Y), float32(q.Z)
	return Mat3{
		1 - 2*y*y - 2*z*z, 2*x*y + 2*w*z, 2*x*z - 2*w*y,
		2*x*y - 2*w*z, 1 - 2*x*x - 2*z*z, 2*y*z + 2*w*x,
		2*x*z + 2*w*y, 2*y*z - 2*w*x, 1 - 2*x*x - 2*y*y,
	}
}

func (qin Quaternion) ToMat4() Mat4 {
	m := qin.RotMat()
	return Mat4{
		m[0], m[1], m[2], 0,
		m[3], m[4], m[5], 0,
		m[6], m[7], m[8], 0,
		0, 0, 0, 1,
	}
}
