package ygl

type FrameBuffer struct {
	Width, Height int
	Fbo           uint32
	Color         uint32
	Depth         uint32
}
