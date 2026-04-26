package yecs

type Spatial struct {
	Buffer         string
	Program        string
	Textures       string
	CurTexture     int
	AssignUniforms func(e EntityId, w *World, cam *Camera, program uint32) error
}

type Sprite struct {
	Pos   [2]int //bottom -left
	Col   int
	Row   int
	Width int
}
