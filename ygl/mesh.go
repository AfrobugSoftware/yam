package ygl

import "github.com/go-gl/gl/v4.3-core/gl"

const (
	INDEX_BUFFER = iota
	POS_M_VB
	TEXCOORDS_M_VB
	NORMAL_M_VB
	WORLD_MAT_M_VP
	NUM_BUFFERS_M
)

type MeshEntry struct {
	NumIndices uint32
	BaseVertex uint32
	BaseIndex  uint32
}

type Mesh struct {
	Vbo         uint32
	Buffers     [NUM_BUFFERS_M]uint32
	Textures    []uint32
	MeshEntries []MeshEntry

	//accumulation buffers
	positions [][3]float32
	normals   [][3]float32
	texCoords [][2]float32
	indices   []uint32
}

func CreateMesh() *Mesh {
	m := &Mesh{}
	gl.CreateVertexArrays(1, &m.Vbo)
	gl.CreateBuffers(int32(len(m.Buffers)), &m.Buffers[0])
	return m
}
