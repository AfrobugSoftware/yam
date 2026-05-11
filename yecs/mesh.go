package yecs

import (
	"bytes"
	"sync"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/qmuntal/gltf"
)

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

const (
	INDEX_M_VB = iota
	POS_M_VB
	NORMAL_M_VB
	TEXCOORDS_M_VB
	NUM_BUFFERS_M
)

type GltfDataFormat struct {
	ComponentType gltf.ComponentType
	ValueType     gltf.AccessorType
	ValueCount    int
}

type Mesh struct {
	MeshId  int
	Vbo     uint32
	Buffers [NUM_BUFFERS_M]uint32
	Formats map[string]GltfDataFormat
	//accumulation buffers
	Positions   *bytes.Buffer
	Normals     *bytes.Buffer
	TexCoords   *bytes.Buffer
	Indices     *bytes.Buffer
	NumVertices uint32
	NumIndices  uint32
}

func CreateMesh() *Mesh {
	m := &Mesh{}
	gl.CreateVertexArrays(1, &m.Vbo)
	gl.CreateBuffers(int32(len(m.Buffers)), &m.Buffers[0])

	m.Positions = bufPool.Get().(*bytes.Buffer)
	m.Positions.Reset()
	m.Normals = bufPool.Get().(*bytes.Buffer)
	m.Normals.Reset()
	m.TexCoords = bufPool.Get().(*bytes.Buffer)
	m.Normals.Reset()
	m.Indices = bufPool.Get().(*bytes.Buffer)
	m.Indices.Reset()

	return m
}

func (m *Mesh) Bind() {
	gl.BindVertexArray(m.Vbo)
}

func (m *Mesh) Unbind() {
	gl.BindVertexArray(0)
}

func (m *Mesh) Setup() {
	if m.Positions.Len() != 0 {
		pos := m.Positions.Bytes()
		//df := m.Formats[gltf.POSITION]
		gl.NamedBufferStorage(m.Buffers[POS_M_VB], m.Positions.Len(), gl.Ptr(&pos[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vbo, 0, 0)
		gl.VertexArrayVertexBuffer(m.Vbo, 0, m.Buffers[POS_M_VB], 0, int32(unsafe.Sizeof(float32(0))*3))
		gl.VertexArrayAttribFormat(m.Vbo, 0, 3, gl.FLOAT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vbo, 0)
	}
	if m.Normals.Len() != 0 {
		normals := m.Normals.Bytes()
		gl.NamedBufferStorage(m.Buffers[NORMAL_M_VB], m.Normals.Len(), gl.Ptr(&normals[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vbo, 1, 1)
		gl.VertexArrayVertexBuffer(m.Vbo, 1, m.Buffers[NORMAL_M_VB], 0, int32(unsafe.Sizeof(float32(0))*3))
		gl.VertexArrayAttribFormat(m.Vbo, 1, 3, gl.FLOAT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vbo, 1)
	}
	if m.TexCoords.Len() != 0 {
		texs := m.TexCoords.Bytes()
		gl.NamedBufferStorage(m.Buffers[TEXCOORDS_M_VB], m.TexCoords.Len(), gl.Ptr(&texs[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vbo, 2, 2)
		gl.VertexArrayVertexBuffer(m.Vbo, 2, m.Buffers[TEXCOORDS_M_VB], 0, int32(unsafe.Sizeof(float32(0))*2))
		gl.VertexArrayAttribFormat(m.Vbo, 2, 2, gl.FLOAT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vbo, 2)
	}

	if m.Indices.Len() != 0 {
		indices := m.Indices.Bytes()
		gl.NamedBufferStorage(m.Buffers[INDEX_M_VB], m.Indices.Len(), gl.Ptr(&indices[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayElementBuffer(m.Vbo, m.Buffers[INDEX_M_VB])
	}
	bufPool.Put(m.Positions)
	bufPool.Put(m.Normals)
	bufPool.Put(m.TexCoords)
	bufPool.Put(m.Indices)
}

func (m *Mesh) ShutDown() {
	gl.DeleteBuffers(NUM_BUFFERS_M, &m.Buffers[0])
	gl.DeleteVertexArrays(1, &m.Vbo)
	clear(m.Formats)
}
