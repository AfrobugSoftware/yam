package ygl

import (
	"math"
	"unsafe"

	gl "github.com/chsc/gogl/gl33"
)

func CreateSphere(sectorCount, stackCount int, radius float64) ([]gl.Float, []uint16, []DataFormat) {
	buffer := make([]gl.Float, 0)
	indices := make([]uint16, 0)
	format := make([]DataFormat, 3)

	sectorStep := 2 * math.Pi / float64(sectorCount)
	stackStep := math.Pi / float64(stackCount)
	lengthInv := 1 / radius

	for i := range stackCount {
		stackAngle := math.Pi/2 - float64(i)*stackStep
		xz := gl.Float(radius * math.Cos(float64(stackAngle)))
		y := gl.Float(radius * math.Sin(float64(stackAngle)))

		k1 := uint16(i * (sectorCount + 1))
		k2 := k1 + uint16(sectorCount+1)
		for j := range sectorCount {
			sectorAngle := float64(j) * sectorStep
			x := xz * gl.Float(math.Sin(sectorAngle))
			z := xz * gl.Float(math.Cos(sectorAngle))

			nx := x * gl.Float(lengthInv)
			ny := y * gl.Float(lengthInv)
			nz := z * gl.Float(lengthInv)

			s := gl.Float(float32(i) / float32(sectorCount))
			t := gl.Float(float32(j) / float32(stackCount))

			buffer = append(buffer, x, y, z, nx, ny, nz, s, t)

			if i != 0 {
				indices = append(indices, k1, k2, k1+1)
			}
			if i != stackCount-1 {
				indices = append(indices, k1+1, k2, k2+1)
			}
			k1++
			k2++
		}
	}
	format[0] = DataFormat{
		Count:  3,
		Stride: int(uintptr(8) * unsafe.Sizeof(gl.Float(0))),
		Offset: 0,
	}
	format[1] = DataFormat{
		Count:  3,
		Stride: int(uintptr(8) * unsafe.Sizeof(gl.Float(0))),
		Offset: 3,
	}
	format[2] = DataFormat{
		Count:  2,
		Stride: int(uintptr(8) * unsafe.Sizeof(gl.Float(0))),
		Offset: 6,
	}

	return buffer, indices, format
}
