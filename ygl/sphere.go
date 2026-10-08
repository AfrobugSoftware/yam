package ygl

import (
	"bytes"
	"encoding/binary"
	"math"
	"reflect"
	"unsafe"
	"yam/y3d"
)

func CreateSphere(sectorCount, stackCount int, radius float64) (dataV, dataI *bytes.Buffer) {
	var count, indx uint32
	verties := make([]y3d.PVertex, 0, stackCount*sectorCount)

	sectorStep := 2 * math.Pi / float32(sectorCount)
	stackStep := math.Pi / float32(stackCount)
	lengthInv := 1 / radius
	dataV = &bytes.Buffer{}
	dataI = &bytes.Buffer{}
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

			verties = append(verties, y3d.PVertex{
				Pos: y3d.Vec3{
					X: x,
					Y: y,
					Z: z,
				},
				Norm: y3d.Vec3{
					X: nx,
					Y: ny,
					Z: nz,
				},
				Tc: y3d.Vec2{
					X: s,
					Y: t,
				},
				Color: [4]uint8{255, 255, 255, 255},
			})
			count++
		}
	}
	addr := unsafe.Pointer(reflect.ValueOf(verties).Index(0).UnsafeAddr())
	b := unsafe.Slice((*byte)(addr), 36*len(verties))
	dataV.Write(b)
	for i := range stackCount {
		k1 := uint32(i * (sectorCount + 1))
		k2 := k1 + uint32(sectorCount) + 1
		for range sectorCount {
			if i != 0 {
				binary.Write(dataI, binary.NativeEndian, k1)
				binary.Write(dataI, binary.NativeEndian, k2)
				binary.Write(dataI, binary.NativeEndian, k1+1)
				indx += 3
			}
			if i != stackCount-1 {
				binary.Write(dataI, binary.NativeEndian, k1+1)
				binary.Write(dataI, binary.NativeEndian, k2)
				binary.Write(dataI, binary.NativeEndian, k2+1)
				indx += 3
			}
			k1++
			k2++
		}
	}
	return
}
