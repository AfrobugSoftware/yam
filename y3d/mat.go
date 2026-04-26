package y3d

import (
	"bytes"
	"fmt"
	"math"
	"text/tabwriter"
)

// Mat4 is a 4x4 column-major matrix (OpenGL convention)
//
//	| m[0]  m[4]  m[8]  m[12] |
//	| m[1]  m[5]  m[9]  m[13] |
//	| m[2]  m[6]  m[10] m[14] |
//	| m[3]  m[7]  m[11] m[15] |
type Mat4 [16]float32
type Mat3 [9]float32
type Mat2 [4]float32

var Identity = Mat4{
	1, 0, 0, 0,
	0, 1, 0, 0,
	0, 0, 1, 0,
	0, 0, 0, 1,
}

func (m Mat4) MulVec4(v Vec4) Vec4 {
	return Vec4{
		X: m[0]*v.X + m[4]*v.Y + m[8]*v.Z + m[12]*v.W,
		Y: m[1]*v.X + m[5]*v.Y + m[9]*v.Z + m[13]*v.W,
		Z: m[2]*v.X + m[6]*v.Y + m[10]*v.Z + m[14]*v.W,
		W: m[3]*v.X + m[7]*v.Y + m[11]*v.Z + m[15]*v.W,
	}
}

func (m Mat4) MulVec3(v Vec3) Vec3 {
	r := m.MulVec4(Vec4{v.X, v.Y, v.Z, 0})
	return Vec3{r.X, r.Y, r.Z}
}

func (a Mat4) Mul(b Mat4) Mat4 {
	var out Mat4
	for col := range 4 {
		for row := range 4 {
			var sum float32
			for k := range 4 {
				sum += a[k*4+row] * b[col*4+k]
			}
			out[col*4+row] = sum
		}
	}
	return out
}

func (m1 *Mat4) MulWith(c float32) {
	m1[0] *= c
	m1[1] *= c
	m1[2] *= c
	m1[3] *= c
	m1[4] *= c
	m1[5] *= c
	m1[6] *= c
	m1[7] *= c
	m1[8] *= c
	m1[9] *= c
	m1[10] *= c
	m1[11] *= c
	m1[12] *= c
	m1[13] *= c
	m1[14] *= c
	m1[15] *= c
}

func (m1 *Mat4) Invert() {
	det := m1.Det()
	if det < NearZero {
		*m1 = Identity //do not know if this is valid
		return
	}

	//m1ake a copy to not override original while reading
	v0 := m1[0]
	v1 := m1[1]
	v2 := m1[2]
	v3 := m1[3]
	v4 := m1[4]
	v5 := m1[5]
	v6 := m1[6]
	v7 := m1[7]
	v8 := m1[8]
	v9 := m1[9]
	v10 := m1[10]
	v11 := m1[11]
	v12 := m1[12]
	v13 := m1[13]
	v14 := m1[14]
	v15 := m1[15]

	//precalculate the most common products
	v7v10 := v7 * v10
	v6v11 := v6 * v11
	v7v9 := v7 * v9
	v5v11 := v5 * v11
	v6v9 := v6 * v9
	v5v10 := v5 * v10
	v1v4 := v1 * v4
	v4v9 := v4 * v9
	v6v8 := v6 * v8
	v5v8 := v5 * v8
	v7v8 := v7 * v8
	v1v12 := v1 * v12
	v2v12 := v2 * v12
	v2v13 := v2 * v13
	v2v15 := v2 * v15
	v3v12 := v3 * v12
	v3v13 := v3 * v13
	v3v14 := v3 * v14
	v4v10 := v4 * v10
	v4v11 := v4 * v11
	v10v15 := v10 * v15
	v7v14 := v7 * v14
	v6v15 := v6 * v15
	v0v11 := v0 * v11
	v1v8 := v1 * v8
	v0v9 := v0 * v9
	v0v5 := v0 * v5
	v0v13 := v0 * v13

	m1[0] = -v7v10*v13 + v6v11*v13 + v7v9*v14 - v5v11*v14 - v6v9*v15 + v5v10*v15
	m1[1] = v3v13*v10 - v2v13*v11 - v3v14*v9 + v1*v11*v14 + v2v15*v9 - v1*v10v15
	m1[2] = -v3v13*v6 + v2v13*v7 + v3v14*v5 - v1*v7v14 - v2v15*v5 + v1*v6v15
	m1[3] = v3*v6v9 - v2*v7v9 - v3*v5v10 + v1*v7v10 + v2*v5v11 - v1*v6v11
	m1[4] = v7v10*v12 - v6v11*v12 - v7v8*v14 + v4v11*v14 + v6v8*v15 - v4v10*v15
	m1[5] = -v3v12*v10 + v2v12*v11 + v3v14*v8 - v0v11*v14 - v2v15*v8 + v0*v10v15
	m1[6] = v3v12*v6 - v2v12*v7 - v3v14*v4 + v0*v7v14 + v2v15*v4 - v0*v6v15
	m1[7] = -v3*v6v8 + v2*v7v8 + v3*v4v10 - v0*v7v10 - v2*v4v11 + v0*v6v11
	m1[8] = -v7v9*v12 + v5v11*v12 + v7v8*v13 - v4v11*v13 - v5v8*v15 + v4v9*v15
	m1[9] = v3v12*v9 - v1v12*v11 - v3v13*v8 + v0v11*v13 + v1v8*v15 - v0v9*v15
	m1[10] = -v3v12*v5 + v1v12*v7 + v3v13*v4 - v0v13*v7 - v1v4*v15 + v0v5*v15
	m1[11] = v3*v5v8 - v1*v7v8 - v3*v4v9 + v0*v7v9 + v1v4*v11 - v0*v5v11
	m1[12] = v6v9*v12 - v5v10*v12 - v6v8*v13 + v4v10*v13 + v5v8*v14 - v4v9*v14
	m1[13] = -v2v12*v9 + v1v12*v10 + v2v13*v8 - v0v13*v10 - v1v8*v14 + v0v9*v14
	m1[14] = v2v12*v5 - v1v12*v6 - v2v13*v4 + v0v13*v6 + v1v4*v14 - v0v5*v14
	m1[15] = -v2*v5v8 + v1*v6v8 + v2*v4v9 - v0*v6v9 - v1v4*v10 + v0*v5v10
	m1.MulWith(1.0 / det)
}

func (m1 *Mat4) Transposed() Mat4 {
	return Mat4{m1[0], m1[4], m1[8], m1[12],
		m1[1], m1[5], m1[9], m1[13],
		m1[2], m1[6], m1[10], m1[14],
		m1[3], m1[7], m1[11], m1[15]}
}

func (m1 *Mat4) Det() float32 {
	return m1[0]*m1[5]*m1[10]*m1[15] - m1[0]*m1[5]*m1[11]*m1[14] - m1[0]*m1[6]*m1[9]*m1[15] + m1[0]*m1[6]*m1[11]*m1[13] + m1[0]*m1[7]*m1[9]*m1[14] - m1[0]*m1[7]*m1[10]*m1[13] - m1[1]*m1[4]*m1[10]*m1[15] + m1[1]*m1[4]*m1[11]*m1[14] + m1[1]*m1[6]*m1[8]*m1[15] - m1[1]*m1[6]*m1[11]*m1[12] - m1[1]*m1[7]*m1[8]*m1[14] + m1[1]*m1[7]*m1[10]*m1[12] + m1[2]*m1[4]*m1[9]*m1[15] - m1[2]*m1[4]*m1[11]*m1[13] - m1[2]*m1[5]*m1[8]*m1[15] + m1[2]*m1[5]*m1[11]*m1[12] + m1[2]*m1[7]*m1[8]*m1[13] - m1[2]*m1[7]*m1[9]*m1[12] - m1[3]*m1[4]*m1[9]*m1[14] + m1[3]*m1[4]*m1[10]*m1[13] + m1[3]*m1[5]*m1[8]*m1[14] - m1[3]*m1[5]*m1[10]*m1[12] - m1[3]*m1[6]*m1[8]*m1[13] + m1[3]*m1[6]*m1[9]*m1[12]
}

func (m1 *Mat4) Transpose() {
	m1[1], m1[2], m1[3], m1[4], m1[6], m1[7], m1[8], m1[9], m1[11], m1[12], m1[13], m1[14] = m1[4], m1[8], m1[12], m1[1], m1[9], m1[13], m1[2], m1[6], m1[14], m1[3], m1[7], m1[11]
}

func Scale(s Vec3) Mat4 {
	return Mat4{
		s.X, 0, 0, 0,
		0, s.Y, 0, 0,
		0, 0, s.Z, 0,
		0, 0, 0, 1,
	}
}

func Translation(t Vec3) Mat4 {
	return Mat4{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		t.X, t.Y, t.Z, 1,
	}
}

func RotationX(angle float64) Mat4 {
	c := float32(math.Cos(angle))
	s := float32(math.Sin(angle))
	return Mat4{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

func RotationY(angle float64) Mat4 {
	c := float32(math.Cos(angle))
	s := float32(math.Sin(angle))
	return Mat4{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

func RotationZ(angle float64) Mat4 {
	c := float32(math.Cos(angle))
	s := float32(math.Sin(angle))
	return Mat4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func RotationAxis(axis Vec3, angle float64) Mat4 {
	c := float32(math.Cos(angle))
	s := float32(math.Sin(angle))
	sum := 1.0 - c

	axis = Normalize(axis)
	var mat Mat4
	// R = I +(sin(a))S + (1-cos(a)S^2)
	mat[0] = (axis.X*axis.X)*sum + c
	mat[1] = (axis.X*axis.Y)*sum - (axis.Z * s)
	mat[2] = (axis.X*axis.Z)*sum + (axis.Y * s)
	mat[3] = 0

	mat[4] = (axis.Y*axis.X)*sum + (axis.Z * s)
	mat[5] = (axis.Y*axis.Y)*sum + c
	mat[6] = (axis.Y*axis.Z)*sum - (axis.X * s)
	mat[7] = 0

	mat[8] = (axis.Z*axis.X)*sum - (axis.Y * s)
	mat[9] = (axis.Z*axis.Y)*sum + (axis.X * s)
	mat[10] = (axis.Z*axis.Z)*sum + c
	mat[11] = 0

	mat[12] = 0
	mat[13] = 0
	mat[14] = 0
	mat[15] = 1.0
	return mat
}

func Ortho(left, right, bottom, top, near, far float32) Mat4 {
	rml := right - left
	tmb := top - bottom
	nmf := near - far

	return Mat4{
		2 / rml, 0, 0, 0,
		0, 2 / tmb, 0, 0,
		0, 0, 2 / nmf, 0,
		(right + left) / (left - right), (top + bottom) / (bottom - top), (far + near) / (far - near), 1,
	}
}

func Frustum(left, right, bottom, top, near, far float32) Mat4 {
	rml := right - left
	tmb := top - bottom
	fmn := far - near

	return Mat4{
		(2 * near) / rml, 0, 0, 0,
		0, (2 * near) / tmb, 0, 0,
		(right + left) / rml, (top + bottom) / tmb, -(far + near) / fmn, -1,
		0, 0, (-2 * far * near) / fmn, 0,
	}
}

func (m1 Mat4) String() string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 4, 4, 1, ' ', tabwriter.AlignRight)
	for i := range 4 {
		for j := range 4 {
			fmt.Fprintf(w, "%f\t", m1[j*4+i])
		}
		fmt.Fprintln(w, "")
	}
	w.Flush()
	return buf.String()
}
