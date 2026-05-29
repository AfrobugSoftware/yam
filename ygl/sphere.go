package ygl

import (
	"encoding/binary"
	"math"
	"yam/yecs"
)

func CreateSphere(sectorCount, stackCount int, radius float64, mesh *yecs.Mesh) yecs.MeshEntry {
	me := yecs.MeshEntry{
		MeshId: mesh.MeshId,
	}
	me.BaseIndex = mesh.NumIndices
	me.BaseVertex = mesh.NumVertices
	var count, indx uint32

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

			binary.Write(mesh.Positions, binary.NativeEndian, x)
			binary.Write(mesh.Positions, binary.NativeEndian, y)
			binary.Write(mesh.Positions, binary.NativeEndian, z)

			binary.Write(mesh.Normals, binary.NativeEndian, nx)
			binary.Write(mesh.Normals, binary.NativeEndian, ny)
			binary.Write(mesh.Normals, binary.NativeEndian, nz)

			binary.Write(mesh.TexCoords, binary.NativeEndian, s)
			binary.Write(mesh.TexCoords, binary.NativeEndian, t)

			count++
		}
	}

	for i := range stackCount {
		k1 := uint32(i * (sectorCount + 1))
		k2 := k1 + uint32(sectorCount) + 1
		for range sectorCount {
			if i != 0 {
				binary.Write(mesh.Indices, binary.NativeEndian, k1)
				binary.Write(mesh.Indices, binary.NativeEndian, k2)
				binary.Write(mesh.Indices, binary.NativeEndian, k1+1)
				indx += 3
			}
			if i != stackCount-1 {
				binary.Write(mesh.Indices, binary.NativeEndian, k1+1)
				binary.Write(mesh.Indices, binary.NativeEndian, k2)
				binary.Write(mesh.Indices, binary.NativeEndian, k2+1)
				indx += 3
			}
			k1++
			k2++
		}
	}
	me.NumIndices = indx
	me.NumVertices = count

	mesh.NumIndices += me.NumIndices
	mesh.NumVertices += me.NumVertices
	return me
}
