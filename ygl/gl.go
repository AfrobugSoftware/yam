package ygl

import (
	"fmt"
	"log"
	"sync"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

func glCheck(tag string) {
	if err := gl.GetError(); err != gl.NO_ERROR {
		log.Fatalf("[GL ERROR] at '%s': 0x%x", tag, err)
	}
}

func glTypeSize(glType int32) int32 {
	switch uint32(glType) {
	case gl.FLOAT, gl.INT, gl.UNSIGNED_INT, gl.BOOL:
		return 4
	case gl.FLOAT_VEC2, gl.INT_VEC2:
		return 8
	case gl.FLOAT_VEC3, gl.INT_VEC3:
		return 12
	case gl.FLOAT_VEC4, gl.INT_VEC4:
		return 16
	case gl.FLOAT_MAT2:
		return 16
	case gl.FLOAT_MAT3:
		return 36
	case gl.FLOAT_MAT4:
		return 64
	default:
		return 0
	}
}

type Gl3 struct {
	Context          sdl.GLContext
	Window           *sdl.Window
	ClearColor       y3d.Vec4
	PixelDepth       uint8
	DoubleBuffer     bool
	MinorVersion     int
	MajorVersion     int
	Height           int32
	Width            int32
	DeferredRenderer *DeferredRenderer
	mu               sync.Mutex
}

func NewYGL(window *sdl.Window, width, height int) (*Gl3, error) {
	g := &Gl3{
		Window: window,
		Height: int32(height),
		Width:  int32(width),
		ClearColor: y3d.Vec4{
			X: 0,
			Y: 0,
			Z: 0,
			W: 1,
		},
	}
	context, err := g.Window.GLCreateContext()
	if err != nil {
		return nil, err
	}
	g.Context = context
	if err = gl.Init(); err != nil {
		panic(err)
	}
	gl.Viewport(0, 0, g.Width, g.Height)
	gl.ClearColor(g.ClearColor.X, g.ClearColor.Y, g.ClearColor.Z, g.ClearColor.W)

	g.DeferredRenderer = CreateDeferredRenderer(int32(width), int32(height), int32(width), int32(height))
	return g, nil
}

func (g *Gl3) ShutDownGL() {
	g.mu.Lock()
	defer g.mu.Unlock()
	sdl.GLDeleteContext(g.Context)
}

func AddProgramSource(vert, frag string) uint32 {

	v, err := CreateShaderFromFile(vert, gl.VERTEX_SHADER)
	if err != nil {
		panic(fmt.Errorf("%s: %v", vert, err))
	}
	f, err := CreateShaderFromFile(frag, gl.FRAGMENT_SHADER)
	if err != nil {
		panic(fmt.Errorf("%s: %v", frag, err))
	}
	p, err := CreateProgram([]uint32{v, f})
	if err != nil {
		panic(err)
	}
	return p
}
