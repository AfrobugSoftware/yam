package ygl

import (
	"math"
	"unsafe"
)

func CreateSphere(sectorCount, stackCount int, radius float64) ([]byte, []uint16, []DataFormat) {
	buffer := make([]float32, 0)
	format := make([]DataFormat, 3)

	sectorStep := 2 * math.Pi / float32(sectorCount)
	stackStep := math.Pi / float32(stackCount)
	lengthInv := 1 / radius

	for i := 0; i <= stackCount; i++ {
		stackAngle := math.Pi/2 - float32(i)*stackStep
		xz := float32(radius * math.Cos(float64(stackAngle)))
		y := float32(radius * math.Sin(float64(stackAngle)))
		for j := 0; j <= sectorCount; j++ {
			sectorAngle := float32(j) * sectorStep
			x := xz * float32(math.Sin(float64(sectorAngle)))
			z := xz * float32(math.Cos(float64(sectorAngle)))

			nx := x * float32(lengthInv)
			ny := y * float32(lengthInv)
			nz := z * float32(lengthInv)

			s := float32(i) / float32(sectorCount)
			t := float32(j) / float32(stackCount)

			buffer = append(buffer, x, y, z, nx, ny, nz, s, t)

		}
	}

	indices := make([]uint16, 0)
	for i := range stackCount {
		k1 := uint16(i * (sectorCount + 1))
		k2 := k1 + uint16(sectorCount) + 1
		for range sectorCount {
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
		Count:         3,
		Stride:        int32(uintptr(8) * unsafe.Sizeof(float32(0))),
		Offset:        0,
		ComponentType: ComponentTypeFloat32,
	}
	format[1] = DataFormat{
		Count:         3,
		Stride:        int32(uintptr(8) * unsafe.Sizeof(float32(0))),
		Offset:        int32(uintptr(3) * unsafe.Sizeof(float32(0))),
		ComponentType: ComponentTypeFloat32,
	}
	format[2] = DataFormat{
		Count:         2,
		Stride:        int32(uintptr(8) * unsafe.Sizeof(float32(0))),
		Offset:        int32(uintptr(6) * unsafe.Sizeof(float32(0))),
		ComponentType: ComponentTypeFloat32,
	}
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&buffer[0])), len(buffer)*int(unsafe.Sizeof(float32(0))))
	return buf, indices, format
}
