package ygl

import (
	"fmt"
	"sync"

	gl "github.com/chsc/gogl/gl33"
	"github.com/veandco/go-sdl2/sdl"
)

type Gl3 struct {
	mu sync.Mutex

	Context      sdl.GLContext
	Window       *sdl.Window
	ClearColor   sdl.Color
	PixelDepth   uint8
	DoubleBuffer bool
	MinorVersion int
	MajorVersion int
	Height       int
	Width        int
	Buffers      map[string]VertBuffer
	Programs     map[string]gl.Uint
	Textures     map[string][]gl.Uint
}

func InitGL(g *Gl3) error {
	context, err := g.Window.GLCreateContext()
	if err != nil {
		return err
	}
	g.Context = context
	gl.Init()
	gl.Viewport(0, 0, gl.Sizei(g.Width), gl.Sizei(g.Height))
	// OPENGL FLAGS
	gl.ClearColor(gl.Float(g.ClearColor.R), gl.Float(g.ClearColor.G), gl.Float(g.ClearColor.B), gl.Float(g.ClearColor.A))
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	return nil
}

func ShutDownGL(g *Gl3) {
	g.mu.Lock()
	defer g.mu.Unlock()
	sdl.GLDeleteContext(g.Context)
}

func (g *Gl3) AddTextures(filePath []string, name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	texs := make([]gl.Uint, 0, len(filePath))
	for i, f := range filePath {
		t, err := CreateTex2D(f, gl.LINEAR, gl.LINEAR)
		if err != nil {
			return err
		}
		texs[i] = t
	}
	g.Textures[name] = texs
	return nil
}

func (g *Gl3) AddPrograms(name string, filePath []string, types []gl.Enum) error {
	if len(filePath) != len(types) {
		return fmt.Errorf("invalid file to types")
	}
	shaders := []gl.Uint{}
	for i, f := range filePath {
		t := types[i]
		s, err := CreateShader(f, t)
		if err != nil {
			return err
		}
		shaders = append(shaders, s)
	}
	p, err := CreateProgram(shaders)
	if err != nil {
		return err
	}
	g.Programs[name] = p
	return nil
}

func (g *Gl3) AddProgramSource(name, vert, frag string) error {
	v, err := CreateShader(vert, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	f, err := CreateShader(frag, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil
	}
	p, err := CreateProgram([]gl.Uint{v, f})
	if err != nil {
		return err
	}
	g.Programs[name] = p
	return nil
}
