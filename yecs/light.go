package yecs

import "yam/y3d"

const (
	POINT_LIGHT = iota
	DIR_LIGHT
	SPOT_LIGHT
	AMBIENT_LIGHT
)

type Light struct {
	Type      int
	Pos       y3d.Vec3
	Intensity float32
	Direction y3d.Vec3
	Diffuse   y3d.Vec3
	Ambient   y3d.Vec3
	Specular  y3d.Vec3
}

func (l *Light) ToUBO() [16]float32 {
	var buf [16]float32
	buf[0] = l.Diffuse.X
	buf[1] = l.Diffuse.Y
	buf[2] = l.Diffuse.Z
	buf[3] = 0.0

	buf[4] = l.Ambient.X
	buf[5] = l.Ambient.Y
	buf[6] = l.Ambient.Z
	buf[7] = 0.0

	buf[8] = l.Specular.X
	buf[9] = l.Specular.Y
	buf[10] = l.Specular.Z
	buf[11] = 0.0

	buf[12] = l.Pos.X
	buf[13] = l.Pos.Y
	buf[14] = l.Pos.Z
	buf[15] = 0.0

	return buf
}
