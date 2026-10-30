package ygl

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"unsafe"
	"yam/y3d"
)

func CreateCube() (dataV, dataI *bytes.Buffer) {
	dataV = &bytes.Buffer{}
	dataI = &bytes.Buffer{}
	verties := make([]y3d.PVertex, 0, 8)
	verties = append(verties,
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: -1, Y: -1, Z: -1,
			},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: -1, Y: 1, Z: -1,
			},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: 1, Y: -1, Z: -1,
			},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: 1, Y: 1, Z: -1,
			},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: 1, Y: -1, Z: 1,
			},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: 1, Y: 1, Z: 1,
			},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: -1, Y: -1, Z: 1,
			},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: -1, Y: 1, Z: 1,
			},
		},
	)

	indices := []uint32{
		0, 1, 2,
		2, 1, 3,
		2, 3, 4,
		4, 3, 5,
		4, 5, 6,
		6, 5, 7,
		6, 7, 0,
		0, 7, 1,
		6, 0, 2,
		2, 4, 6,
		7, 5, 3,
		7, 3, 1,
	}
	for i := 0; i < len(indices); i += 6 {
		a, b, c := verties[indices[i]].Pos, verties[indices[i+1]].Pos,
			verties[indices[i+2]].Pos
		n := y3d.Normalize(y3d.Cross(y3d.Sub(b, a), y3d.Sub(c, a)))
		verties[indices[i]].Norm = y3d.Add(verties[indices[i]].Norm, y3d.Smul(n, 1.0/3.0))
		verties[indices[i+1]].Norm = y3d.Normalize(y3d.Add(verties[indices[i+1]].Norm, y3d.Smul(n, 1.0/3.0)))
		verties[indices[i+2]].Norm = y3d.Normalize(y3d.Add(verties[indices[i+2]].Norm, y3d.Smul(n, 1.0/3.0)))
		verties[indices[i+3]].Norm = y3d.Normalize(y3d.Add(verties[indices[i+3]].Norm, y3d.Smul(n, 1.0/3.0)))
		verties[indices[i+4]].Norm = y3d.Normalize(y3d.Add(verties[indices[i+4]].Norm, y3d.Smul(n, 1.0/3.0)))
		verties[indices[i+5]].Norm = y3d.Normalize(y3d.Add(verties[indices[i+5]].Norm, y3d.Smul(n, 1.0/3.0)))
	}

	addr := unsafe.Pointer(reflect.ValueOf(verties).Index(0).UnsafeAddr())
	b := unsafe.Slice((*byte)(addr), 36*len(verties))
	dataV.Write(b)
	binary.Write(dataI, binary.NativeEndian, indices)
	return
}

func CreateQuad() (dataV, dataI *bytes.Buffer) {
	dataV = &bytes.Buffer{}
	dataI = &bytes.Buffer{}
	verties := make([]y3d.PVertex, 0, 4)
	indices := []uint32{
		0, 1, 2,
		2, 1, 3,
	}
	verties = append(verties,
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: -1.00, Y: -1.00, Z: 0,
			},
			Norm: y3d.Vec3{Z: 1.0},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: -1.00, Y: 1.00, Z: 0,
			},
			Norm: y3d.Vec3{Z: 1.0},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: 1.00, Y: -1.00, Z: 0,
			},
			Norm: y3d.Vec3{Z: 1.0},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: 1.00, Y: 1.00, Z: 0,
			},
			Norm: y3d.Vec3{Z: 1.0},
		},
	)
	addr := unsafe.Pointer(reflect.ValueOf(verties).Index(0).UnsafeAddr())
	b := unsafe.Slice((*byte)(addr), 36*len(verties))
	dataV.Write(b)

	binary.Write(dataI, binary.NativeEndian, indices)
	return
}
func CreateYPlaneQuad() (dataV, dataI *bytes.Buffer) {
	dataV = &bytes.Buffer{}
	dataI = &bytes.Buffer{}
	verties := make([]y3d.PVertex, 0, 4)
	indices := []uint32{
		0, 1, 2,
		2, 1, 3,
	}
	verties = append(verties,
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: -1.00, Z: -1.00,
			},
			Norm: y3d.Vec3{Y: 1.0},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: -1.00, Z: 1.00,
			},
			Norm: y3d.Vec3{Y: 1.0},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: 1.00, Z: -1.00,
			},
			Norm: y3d.Vec3{Y: 1.0},
		},
		y3d.PVertex{
			Pos: y3d.Vec3{
				X: 1.00, Z: 1.00,
			},
			Norm: y3d.Vec3{Y: 1.0},
		},
	)
	addr := unsafe.Pointer(reflect.ValueOf(verties).Index(0).UnsafeAddr())
	b := unsafe.Slice((*byte)(addr), 36*len(verties))
	dataV.Write(b)

	binary.Write(dataI, binary.NativeEndian, indices)
	return
}
