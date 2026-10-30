package y3d

import (
	"math"
)

// New returns a new quaternion
func NewQuaternion(w, x, y, z float32) Quaternion {
	return Quaternion{W: w, X: x, Y: y, Z: z}
}

func IdenQuat() Quaternion {
	return Quaternion{W: 1}
}

func Pure(x, y, z float32) Quaternion {
	return Quaternion{X: x, Y: y, Z: z}
}

// A quternion
type Quaternion struct {
	W, X, Y, Z float32
}

func (qin Quaternion) Conj() Quaternion {
	qin.X = -qin.X
	qin.Y = -qin.Y
	qin.Z = -qin.Z
	return qin
}

func (qin Quaternion) Norm2() float32 {
	return qin.W*qin.W + qin.X*qin.X + qin.Y*qin.Y + qin.Z*qin.Z
}

func (qin Quaternion) Neg() Quaternion {
	qin.W = -qin.W
	qin.X = -qin.X
	qin.Y = -qin.Y
	qin.Z = -qin.Z
	return qin
}

func (qin Quaternion) Norm() float32 {
	return float32(math.Sqrt(float64(qin.Norm2())))
}

func Scalar(w float32) Quaternion {
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

// a must be in radians
func FromAngleAxis(axis Vec3, a float32) Quaternion {
	axis = Normalize(axis)
	s := float32(math.Sin(float64(a / 2)))
	w := float32(math.Cos(float64(a / 2)))
	x := axis.X * s
	y := axis.Y * s
	z := axis.Z * s
	return NewQuaternion(w, x, y, z)
}

func ProdQuaternion(qin ...Quaternion) Quaternion {
	qout := Quaternion{1, 0, 0, 0}
	var w, x, y, z float32
	for _, q := range qin {
		w = qout.W*q.W - qout.X*q.X - qout.Y*q.Y - qout.Z*q.Z
		x = qout.W*q.X + qout.X*q.W + qout.Y*q.Z - qout.Z*q.Y
		y = qout.W*q.Y + qout.Y*q.W + qout.Z*q.X - qout.X*q.Z
		z = qout.W*q.Z + qout.Z*q.W + qout.X*q.Y - qout.Y*q.X
		qout = Quaternion{w, x, y, z}
	}
	return qout
}

func (q Quaternion) Unit() Quaternion {
	mag := float32(math.Sqrt(float64(q.W*q.W + q.X*q.X + q.Y*q.Y + q.Z*q.Z)))
	if mag == 0 {
		return Quaternion{W: 1} // identity, avoid divide by zero
	}

	return Quaternion{
		W: q.W / mag,
		X: q.X / mag,
		Y: q.Y / mag,
		Z: q.Z / mag,
	}
}

func (qin Quaternion) Inv() Quaternion {
	k2 := qin.Norm2()
	q := qin.Conj()
	return Quaternion{q.W / k2, q.X / k2, q.Y / k2, q.Z / k2}
}

func (qin Quaternion) Rotate(vec Vec3) Vec3 {
	u := Vec3{X: float32(qin.X), Y: float32(qin.Y), Z: float32(qin.Z)}
	s := float32(qin.W)

	term1 := Smul(u, 2*Dot(u, vec))
	term2 := Smul(vec, s*s-Dot(u, u))
	term3 := Smul(Cross(u, vec), 2*s)
	return Add(Add(term2, term1), term3)
}

func (q Quaternion) Euler() (float32, float32, float32) {
	r := q.Unit()
	phi := math.Atan2(float64(2*(r.W*r.X+r.Y*r.Z)), float64(1-2*(r.X*r.X+r.Y*r.Y)))
	theta := math.Asin(float64(2 * (r.W*r.Y - r.Z*r.X)))
	psi := math.Atan2(float64(2*(r.X*r.Y+r.W*r.Z)), float64(1-2*(r.Y*r.Y+r.Z*r.Z)))
	return float32(phi), float32(theta), float32(psi)
}

func FromEuler(phi, theta, psi float32) Quaternion {
	cp, sp := math.Cos(float64(phi/2)), math.Sin(float64(phi/2))
	ct, st := math.Cos(float64(theta/2)), math.Sin(float64(theta/2))
	cs, ss := math.Cos(float64(psi/2)), math.Sin(float64(psi/2))
	return Quaternion{
		W: float32(cp*ct*cs + sp*st*ss),
		X: float32(sp*ct*cs - cp*st*ss),
		Y: float32(cp*st*cs + sp*ct*ss),
		Z: float32(cp*ct*ss - sp*st*cs),
	}
}

func (qin Quaternion) RotMat() Mat3 {
	q := qin.Unit()
	xx, yy, zz := q.X*q.X, q.Y*q.Y, q.Z*q.Z
	xy, xz, yz := q.X*q.Y, q.X*q.Z, q.Y*q.Z
	wx, wy, wz := q.W*q.X, q.W*q.Y, q.W*q.Z
	return Mat3{
		float32(1 - 2*(yy+zz)), float32(2 * (xy - wz)), float32(2 * (wy + xz)),
		float32(2 * (wz + xy)), float32(1 - 2*(zz+xx)), float32(2 * (yz - wx)),
		float32(2 * (xz - wy)), float32(2 * (wx + yz)), float32(1 - 2*(xx+yy)),
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

func Slerp(a, b Quaternion, t float32) Quaternion {
	dot := a.W*b.W + a.X*b.X + a.Y*b.Y + a.Z*b.Z
	const threshold = 0.9995
	if math.Abs(float64(dot)) > threshold {
		q := Quaternion{
			X: a.X*(1-t) + b.X*t,
			Y: a.Y*(1-t) + b.Y*t,
			Z: a.Z*(1-t) + b.Z*t,
			W: a.W*(1-t) + b.W*t,
		}
		return q.Unit()
	}
	d := math.Acos(float64(dot))
	sinD := math.Sin(d)
	ooSinD := 1 / sinD
	t1 := float32(math.Sin(float64(float32(d)*(1-t))) * ooSinD)
	t2 := float32(math.Sin(d*float64(t)) * ooSinD)

	return Quaternion{
		X: a.X*t1 + b.X*t2,
		Y: a.Y*t1 + b.Y*t2,
		Z: a.Z*t1 + b.Z*t2,
		W: a.W*t1 + b.W*t2,
	}
}
