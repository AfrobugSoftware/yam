package ycore

import (
	"bytes"
	"encoding/gob"
	"yam/y3d"
)

type Face struct {
	Indices    [3]uint32
	Normal     y3d.Vec3
	MeshId     int
	MaterialId int
	Flag       uint8
}

type Mesh struct {
	Name        string
	NumFaces    uint32
	MaterialId  int
	FaceIndices []int
}

type MMaterial struct {
	Name          string
	Ambient       [4]float32
	Diffuse       [4]float32
	Specular      [4]float32
	Emissive      [4]float32
	SpecularPower float32
	Transparency  float32
	Textures      []string
	Flag          uint8
}

type Model struct {
	Vertices    *bytes.Buffer
	VertexType  string
	Materials   []MMaterial
	Faces       []Face
	Meshes      []Mesh
	Joints      []Joint
	Name        string
	Version     string
	Type        string
	NumVertices uint32
	NumIndices  uint32
	NumFaces    uint32
	NumMesh     uint32
	NumJoints   uint32
}

func (m *Model) Write(e *gob.Encoder) error {
	return e.Encode(*m)
}

func (m *Model) Read(d *gob.Decoder) error {
	return d.Decode(m)
}
