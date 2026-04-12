package y3d

import (
	"fmt"
	"math"
)

var (
	UNIT_X = Vec3{X: 1.0, Y: 0.0, Z: 0.0}
	UNIT_Y = Vec3{X: 0.0, Y: 1.0, Z: 0.0}
	UNIT_Z = Vec3{X: 0.0, Y: 0.0, Z: 1.0}
	ZEROV  = Vec3{}
)

type Vec3 struct {
	X, Y, Z float32
}

func (v Vec3) ToSlice() []float32 {
	return []float32{
		v.X, v.Y, v.Z,
	}
}

type IVec3 struct {
	X, Y, Z int
}

func Add(lhs, rhs Vec3) Vec3 {
	return Vec3{
		X: lhs.X + rhs.X,
		Y: lhs.Y + rhs.Y,
		Z: lhs.Z + rhs.Z,
	}
}

func (v Vec3) Equal(q Vec3) bool {
	return v.X == q.X && v.Y == q.Y && v.Z == q.Z
}

func Sub(lhs, rhs Vec3) Vec3 {
	return Vec3{
		X: lhs.X - rhs.X,
		Y: lhs.Y - rhs.Y,
		Z: lhs.Z - rhs.Z,
	}
}

func Smul(v Vec3, s float32) Vec3 {
	return Vec3{
		X: v.X * s,
		Y: v.Y * s,
		Z: v.Z * s,
	}
}

func NegateVec3(v Vec3) Vec3 {
	return Vec3{
		X: -v.X,
		Y: -v.Y,
		Z: -v.Z,
	}
}

func Mul(lhs, rhs Vec3) Vec3 {
	return Vec3{
		X: lhs.X * rhs.X,
		Y: lhs.Y * rhs.Y,
		Z: lhs.Z * rhs.Z,
	}
}

func (v Vec3) Length() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y + v.Z*v.Z)))
}

func (v Vec3) LengthSq() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func Normalize(v Vec3) Vec3 {
	m := float32(v.Length())
	if m > 0 {
		return Vec3{
			X: v.X / m,
			Y: v.Y / m,
			Z: v.Z / m,
		}
	}
	return Vec3{X: 0, Y: 0, Z: 0}
}

func Distance(v1, v2 Vec3) float32 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	dz := v1.Z - v2.Z
	return float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
}

func DistanceSqured(v1, v2 Vec3) float32 {
	dx := v1.X - v2.X
	dy := v1.Y - v2.Y
	dz := v1.Z - v2.Z
	return dx*dx + dy*dy + dz*dz
}

func GetForward2D(rad float64) Vec3 {
	x := float32(math.Cos(rad))
	y := float32(-math.Sin(rad)) //sdl +y is down
	return Vec3{
		X: x,
		Y: y,
		Z: 0,
	}
}

func GetAngle2D(v Vec3) float64 {
	return math.Atan2(float64(-v.Y), float64(v.X))
}

func Dot(v, q Vec3) float32 {
	return v.X*q.X + v.Y*q.Y + v.Z*q.Z
}

func Cross(v, q Vec3) Vec3 {
	x := v.Y*q.Z - v.Z*q.Y
	y := v.Z*q.X - v.X*q.Z
	z := v.X*q.Y - v.Y*q.X
	return Vec3{
		X: x,
		Y: y,
		Z: z,
	}
}

func Lerp(a, b Vec3, t float32) Vec3 {
	t1 := 1 - t
	return Vec3{
		X: a.X*t1 + b.X*t,
		Y: a.Y*t1 + b.Y*t,
		Z: a.Z*t1 + b.Z*t,
	}
}

// assumes N is normalised
func ReflectVec3(I Vec3, N Vec3) Vec3 {
	r := Sub(I, Smul(N, 2*Dot(I, N)))
	return r
}

func (v Vec3) String() string {
	return fmt.Sprintf("[X:%.2f Y:%.2f Z:%.2f]\n", v.X, v.Y, v.Z)
}
