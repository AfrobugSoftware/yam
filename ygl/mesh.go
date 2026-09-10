package ygl

import (
	"bytes"
	"sync"
	"unsafe"

	"github.com/go-gl/gl/v4.3-core/gl"
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
	TEXCOORDS2_M_VB
	JOINTS_M_VB
	WEIGHTS_M_VB
	NUM_BUFFERS_M
)

type Mesh struct {
	MeshId     int
	Vao        uint32
	UseIndices bool
	Buffers    [NUM_BUFFERS_M]uint32
	//accumulation buffers
	Positions    *bytes.Buffer
	Normals      *bytes.Buffer
	TexCoords    *bytes.Buffer
	TexCoords2   *bytes.Buffer
	Indices      *bytes.Buffer
	JointIndices *bytes.Buffer
	WeightValues *bytes.Buffer
	NumVertices  uint32
	NumIndices   uint32
}

func CreateMesh() *Mesh {
	m := &Mesh{}
	gl.CreateVertexArrays(1, &m.Vao)
	gl.CreateBuffers(int32(len(m.Buffers)), &m.Buffers[0])

	m.Positions = bufPool.Get().(*bytes.Buffer)
	m.Positions.Reset()
	m.Normals = bufPool.Get().(*bytes.Buffer)
	m.Normals.Reset()
	m.TexCoords = bufPool.Get().(*bytes.Buffer)
	m.TexCoords.Reset()
	m.Indices = bufPool.Get().(*bytes.Buffer)
	m.Indices.Reset()
	m.JointIndices = bufPool.Get().(*bytes.Buffer)
	m.JointIndices.Reset()
	m.WeightValues = bufPool.Get().(*bytes.Buffer)
	m.WeightValues.Reset()
	m.TexCoords2 = bufPool.Get().(*bytes.Buffer)
	m.TexCoords2.Reset()

	return m
}

func (m *Mesh) Bind() {
	gl.BindVertexArray(m.Vao)
}

func (m *Mesh) Unbind() {
	gl.BindVertexArray(0)
}

func (m *Mesh) Setup() {
	if m.Positions.Len() != 0 {
		pos := m.Positions.Bytes()
		gl.NamedBufferStorage(m.Buffers[POS_M_VB], m.Positions.Len(), gl.Ptr(&pos[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vao, 0, 0)
		gl.VertexArrayVertexBuffer(m.Vao, 0, m.Buffers[POS_M_VB], 0, int32(unsafe.Sizeof(float32(0))*3))
		gl.VertexArrayAttribFormat(m.Vao, 0, 3, gl.FLOAT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vao, 0)
	}
	if m.Normals.Len() != 0 {
		normals := m.Normals.Bytes()
		gl.NamedBufferStorage(m.Buffers[NORMAL_M_VB], m.Normals.Len(), gl.Ptr(&normals[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vao, 1, 1)
		gl.VertexArrayVertexBuffer(m.Vao, 1, m.Buffers[NORMAL_M_VB], 0, int32(unsafe.Sizeof(float32(0))*3))
		gl.VertexArrayAttribFormat(m.Vao, 1, 3, gl.FLOAT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vao, 1)
	}
	if m.TexCoords.Len() != 0 {
		texs := m.TexCoords.Bytes()
		gl.NamedBufferStorage(m.Buffers[TEXCOORDS_M_VB], m.TexCoords.Len(), gl.Ptr(&texs[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vao, 2, 2)
		gl.VertexArrayVertexBuffer(m.Vao, 2, m.Buffers[TEXCOORDS_M_VB], 0, int32(unsafe.Sizeof(float32(0))*2))
		gl.VertexArrayAttribFormat(m.Vao, 2, 2, gl.FLOAT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vao, 2)
	}
	if m.TexCoords2.Len() != 0 {
		texs := m.TexCoords2.Bytes()
		gl.NamedBufferStorage(m.Buffers[TEXCOORDS2_M_VB], m.TexCoords2.Len(), gl.Ptr(&texs[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vao, 3, 3)
		gl.VertexArrayVertexBuffer(m.Vao, 3, m.Buffers[TEXCOORDS2_M_VB], 0, int32(unsafe.Sizeof(float32(0))*2))
		gl.VertexArrayAttribFormat(m.Vao, 3, 2, gl.FLOAT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vao, 3)
	}
	if m.JointIndices.Len() != 0 {
		joints := m.JointIndices.Bytes()
		gl.NamedBufferStorage(m.Buffers[JOINTS_M_VB], m.JointIndices.Len(), gl.Ptr(&joints[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vao, 4, 4)
		gl.VertexArrayVertexBuffer(m.Vao, 4, m.Buffers[JOINTS_M_VB], 0, int32(unsafe.Sizeof(uint16(0))*4))
		gl.VertexArrayAttribFormat(m.Vao, 4, 4, gl.UNSIGNED_SHORT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vao, 4)
	}
	if m.WeightValues.Len() != 0 {
		weights := m.WeightValues.Bytes()
		gl.NamedBufferStorage(m.Buffers[WEIGHTS_M_VB], m.WeightValues.Len(), gl.Ptr(&weights[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayAttribBinding(m.Vao, 5, 5)
		gl.VertexArrayVertexBuffer(m.Vao, 5, m.Buffers[WEIGHTS_M_VB], 0, int32(unsafe.Sizeof(float32(0))*4))
		gl.VertexArrayAttribFormat(m.Vao, 5, 4, gl.FLOAT, false, 0)
		gl.EnableVertexArrayAttrib(m.Vao, 5)
	}
	if m.Indices.Len() != 0 {
		indices := m.Indices.Bytes()
		gl.NamedBufferStorage(m.Buffers[INDEX_M_VB], m.Indices.Len(), gl.Ptr(&indices[0]), gl.DYNAMIC_STORAGE_BIT)
		gl.VertexArrayElementBuffer(m.Vao, m.Buffers[INDEX_M_VB])
	}
	bufPool.Put(m.Positions)
	bufPool.Put(m.Normals)
	bufPool.Put(m.TexCoords)
	bufPool.Put(m.Indices)
	bufPool.Put(m.JointIndices)
	bufPool.Put(m.WeightValues)
	bufPool.Put(m.TexCoords2)
}

func (m *Mesh) ShutDown() {
	gl.DeleteBuffers(NUM_BUFFERS_M, &m.Buffers[0])
	gl.DeleteVertexArrays(1, &m.Vao)
}
