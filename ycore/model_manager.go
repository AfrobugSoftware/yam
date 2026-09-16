package ycore

import "bytes"

type ModelData struct {
	Vbuf       *bytes.Buffer
	Ibuf       *bytes.Buffer
	VertexType string
	Skin       int
}

type ModelManager struct {
	Models map[string]ModelData
}
