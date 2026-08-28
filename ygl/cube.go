package ygl

import (
	"encoding/binary"
	"yam/yecs"
)

func CreateCube(mesh *yecs.Mesh) yecs.MeshEntry {
	me := yecs.MeshEntry{
		MeshId:      mesh.MeshId,
		BaseVertex:  mesh.NumVertices,
		BaseIndex:   mesh.NumIndices,
		NumIndices:  36,
		NumVertices: 8,
	}
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
	binary.Write(mesh.Positions, binary.NativeEndian, buf)
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
	binary.Write(mesh.Indices, binary.NativeEndian, indices)
	mesh.NumIndices += me.NumIndices
	mesh.NumVertices += me.NumVertices
	return me
}
