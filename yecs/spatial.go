package yecs

type Spatial struct {
	Program   uint32
	Textures  uint32
	Buf       uint32
	Indx      uint32
	VertArray uint32
	IndxCount int32

	CurTexture     int
	AssignUniforms func(e EntityId, w *World, cam *Camera, program uint32) error
}

type Sprite struct {
	Pos   [2]int //bottom -left
	Col   int
	Row   int
	Width int
}
