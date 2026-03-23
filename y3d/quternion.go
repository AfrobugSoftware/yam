package y3d

import (
	"math"
)

// New returns a new quaternion
func NewQuaternion(w, x, y, z float64) Quaternion {
	return Quaternion{W: w, X: x, Y: y, Z: z}
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
func (qin Quaternion) RotMat() [9]float32 {
	q := qin.Unit()
	m := [9]float32{}
	m[0] = float32(1 - 2*(q.Y*q.Y+q.Z*q.Z))
	m[1] = float32(2 * (q.X*q.Y - q.W*q.Z))
	m[2] = float32(2 * (q.W*q.Y + q.X*q.Z))

	m[3] = float32(1 - 2*(q.Z*q.Z+q.X*q.X))
	m[4] = float32(2 * (q.Y*q.Z - q.W*q.X))
	m[5] = float32(2 * (q.W*q.Z + q.Y*q.X))

	m[6] = float32(1 - 2*(q.X*q.X+q.Y*q.Y))
	m[7] = float32(2 * (q.Z*q.X - q.W*q.Y))
	m[8] = float32(2 * (q.W*q.X + q.Z*q.Y))
	return m
}

func (qin Quaternion) ToMat4() Mat4 {
	m := qin.RotMat()
	return Mat4{
		m[0], m[1], m[2], 0,
		m[3], m[4], m[5], 0,
		m[6], m[7], m[8], 0,
		0, 0, 0, 0,
	}
}
