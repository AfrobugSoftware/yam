package ygl

import (
	"bytes"
	"encoding/binary"
)

func CreateCube() (dataV, dataI *bytes.Buffer) {
	dataV = &bytes.Buffer{}
	dataI = &bytes.Buffer{}
	buf := []float32{
		-0.25, -0.25, -0.25,
		-0.25, 0.25, -0.25,
		0.25, -0.25, -0.25,
		0.25, 0.25, -0.25,
		0.25, -0.25, 0.25,
		0.25, 0.25, 0.25,
		-0.25, -0.25, 0.25,
		-0.25, 0.25, 0.25,
	}
	binary.Write(dataV, binary.NativeEndian, buf)
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
	binary.Write(dataI, binary.NativeEndian, indices)
	return
}

func CreateQuad() (dataV, dataI *bytes.Buffer) {
	dataV = &bytes.Buffer{}
	dataI = &bytes.Buffer{}
	buf := []float32{
		-1.00, -1.00, -0.25,
		-1.00, 1.00, -0.25,
		1.00, -1.00, -0.25,
		1.00, 1.00, -0.25,
	}
	binary.Write(dataV, binary.NativeEndian, buf)
	indices := []uint32{
		0, 1, 2,
		2, 1, 3,
	}
	binary.Write(dataI, binary.NativeEndian, indices)
	return
}
