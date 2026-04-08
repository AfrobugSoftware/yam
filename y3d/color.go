package y3d

type Color [4]float32

func (c Color) AsBytes() [4]uint8 {
	return [4]uint8{
		uint8(c[0] * 255),
		uint8(c[1] * 255),
		uint8(c[2] * 255),
		uint8(c[3] * 255),
	}
}

func FromBytes(r, g, b, a uint8) Color {
	return Color{
		float32(r / 255),
		float32(b / 255),
		float32(g / 255),
		float32(a / 255),
	}
}

func (c Color) AsVec3() Vec3 {
	return Vec3{
		X: c[0],
		Y: c[1],
		Z: c[2],
	}
}
