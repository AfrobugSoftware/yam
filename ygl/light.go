package ygl

import (
	"encoding/binary"
	"io"
	"yam/y3d"
)

const (
	POINT_LIGHT = iota
	DIR_LIGHT
	SPOT_LIGHT
	AMBIENT_LIGHT
)

type Light struct {
	Pos         y3d.Vec4
	Direction   y3d.Vec4
	Diffuse     y3d.Vec4
	Ambient     y3d.Vec4
	Specular    y3d.Vec4
	Attenuation y3d.Vec4
	Type        int
	Intensity   float32
	Range       float32
	FallOff     float32
}

func (l *Light) Write(b io.Writer) {
	binary.Write(b, binary.NativeEndian, l.Pos.ToSlice())
	binary.Write(b, binary.NativeEndian, l.Direction.ToSlice())
	binary.Write(b, binary.NativeEndian, l.Diffuse.ToSlice())
	binary.Write(b, binary.NativeEndian, l.Ambient.ToSlice())
	binary.Write(b, binary.NativeEndian, l.Specular.ToSlice())
	binary.Write(b, binary.NativeEndian, l.Attenuation.ToSlice())
	binary.Write(b, binary.NativeEndian, l.Type)
	binary.Write(b, binary.NativeEndian, l.Intensity)
	binary.Write(b, binary.NativeEndian, l.Range)
	binary.Write(b, binary.NativeEndian, l.FallOff)
}
