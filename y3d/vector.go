package y3d

type Vec3 struct {
	X, Y, Z float64
}

func Add(lhs, rhs Vec3) Vec3 {
	return Vec3{
		X: lhs.X + rhs.X,
		Y: lhs.Y + rhs.Y,
		Z: lhs.Z + rhs.Z,
	}
}

func Sub(lhs, rhs Vec3) Vec3 {
	return Vec3{
		X: lhs.X - rhs.X,
		Y: lhs.Y - rhs.Y,
		Z: lhs.Z - rhs.Z,
	}
}

func Smul(v Vec3, s float64) Vec3 {
	return Vec3{
		X: v.X * s,
		Y: v.Y * s,
		Z: v.Z * s,
	}
}

func Mul(lhs, rhs Vec3) Vec3 {
	return Vec3{
		X: lhs.X * rhs.X,
		Y: lhs.Y * rhs.Y,
		Z: lhs.Z * rhs.Z,
	}
}
