package y3d

import "math"

// Mat4 is a 4x4 column-major matrix (OpenGL convention)
//
//	| m[0]  m[4]  m[8]  m[12] |
//	| m[1]  m[5]  m[9]  m[13] |
//	| m[2]  m[6]  m[10] m[14] |
//	| m[3]  m[7]  m[11] m[15] |
type Mat4 [16]float64

type Vec4 struct{ X, Y, Z, W float64 }

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
	r := m.MulVec4(Vec4{v.X, v.Y, v.Z, 1})
	return Vec3{r.X, r.Y, r.Z}
}

func (a Mat4) Mul(b Mat4) Mat4 {
	var out Mat4
	for col := range 4 {
		for row := range 4 {
			var sum float64
			for k := range 4 {
				sum += a[k*4+row] * b[col*4+k]
			}
			out[col*4+row] = sum
		}
	}
	return out
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
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		1, 0, 0, 0,
		0, c, s, 0,
		0, -s, c, 0,
		0, 0, 0, 1,
	}
}

func RotationY(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		c, 0, -s, 0,
		0, 1, 0, 0,
		s, 0, c, 0,
		0, 0, 0, 1,
	}
}

func RotationZ(angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
	return Mat4{
		c, s, 0, 0,
		-s, c, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

func RotationAxis(axis Vec3, angle float64) Mat4 {
	c := math.Cos(angle)
	s := math.Sin(angle)
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

func Ortho(left, right, bottom, top, near, far float64) Mat4 {
	rml := right - left
	tmb := top - bottom
	fmn := far - near

	return Mat4{
		2 / rml, 0, 0, 0,
		0, 2 / tmb, 0, 0,
		0, 0, -2 / fmn, 0,
		-(right + left) / rml, -(top + bottom) / tmb, -(far + near) / fmn, 1,
	}
}
