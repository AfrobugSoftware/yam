package ygl

import (
	"fmt"
	"log"
	"sync"
	"yam/yecs"

	gl "github.com/chsc/gogl/gl33"
	"github.com/veandco/go-sdl2/sdl"
)

type Gl3 struct {
	Context      sdl.GLContext
	Window       *sdl.Window
	ClearColor   sdl.Color
	PixelDepth   uint8
	DoubleBuffer bool
	MinorVersion int
	MajorVersion int
	Height       int
	Width        int

	mu       sync.Mutex
	buffers  map[string]VertBuffer
	programs map[string]gl.Uint
	textures map[string][]gl.Uint
}

func NewYGL(window *sdl.Window, width, height int) (*Gl3, error) {
	g := &Gl3{
		Window:   window,
		buffers:  make(map[string]VertBuffer),
		programs: make(map[string]gl.Uint),
		textures: make(map[string][]gl.Uint),
		Height:   height,
		Width:    width,
		ClearColor: sdl.Color{
			R: 0,
			G: 0,
			B: 0,
		},
	}
	context, err := g.Window.GLCreateContext()
	if err != nil {
		return nil, err
	}
	g.Context = context
	gl.Init()
	gl.Viewport(0, 0, gl.Sizei(g.Width), gl.Sizei(g.Height))
	gl.ClearColor(gl.Float(g.ClearColor.R/255), gl.Float(g.ClearColor.G/255), gl.Float(g.ClearColor.B/255), gl.Float(g.ClearColor.A/255))
	return g, nil
}

func (g *Gl3) ShutDownGL() {
	g.mu.Lock()
	defer g.mu.Unlock()
	sdl.GLDeleteContext(g.Context)
}

func (g *Gl3) AddTextures(filePath []string, name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	texs := make([]gl.Uint, 0, len(filePath))
	for _, f := range filePath {
		t, err := CreateTex2D(f, gl.LINEAR, gl.LINEAR)
		if err != nil {
			return err
		}
		texs = append(texs, t)
	}
	g.textures[name] = texs
	return nil
}

func (g *Gl3) AddPrograms(name string, filePath []string, types []gl.Enum) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(filePath) != len(types) {
		return fmt.Errorf("invalid file to types")
	}
	shaders := []gl.Uint{}
	for i, f := range filePath {
		t := types[i]
		s, err := CreateShaderFromFile(f, t)
		if err != nil {
			return err
		}
		shaders = append(shaders, s)
	}
	p, err := CreateProgram(shaders)
	if err != nil {
		return err
	}
	g.programs[name] = p
	return nil
}

func (g *Gl3) AddProgramSource(name, vert, frag string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

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
	g.programs[name] = p
	return nil
}

func (g *Gl3) AddSprite(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	vbuf := CreateVextexBuffer(SpriteData[:], SpriteIndices, SpriteFormat[:])
	g.buffers[name] = vbuf
}

func (g *Gl3) AddVertexBuffer(name string, data []gl.Float, indx []uint16, formats []DataFormat) {
	g.mu.Lock()
	defer g.mu.Unlock()

	vbuf := CreateVextexBuffer(data, indx, formats)
	g.buffers[name] = vbuf
}

func (g *Gl3) DrawSprites(w *yecs.World) {
	//now := time.Now()
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	camera := w.Query([]yecs.ComponentId{yecs.CameraComponent})
	sprites := w.Query([]yecs.ComponentId{yecs.SpatialComponent, yecs.TransformComponent, yecs.RenderStateComponent})
	if len(camera) == 0 {
		log.Println("no camera attached to scene")
		return
	}
	MainCam := w.GetComponent(camera[0], yecs.CameraComponent).(yecs.Camera)
	view := MainCam.GetViewTransformation()
	proj := MainCam.GetProjectionTransformation()
	projView := proj.Mul(view)
	var curBuf, curProgram, curTexture string
	var program gl.Uint
	var drawBuffer VertBuffer
	for _, e := range sprites {
		s := w.GetComponent(e, yecs.SpatialComponent).(yecs.Spatial)
		if w.HasComponent(e, yecs.BoxComponent) {
			box := w.GetComponent(e, yecs.BoxComponent).(yecs.Box)
			if MainCam.CullView(box, projView) {
				continue
			}
		}
		r := w.GetComponent(e, yecs.RenderStateComponent).(yecs.RenderState)
		if curBuf != s.Buffer {
			b, ok := g.buffers[s.Buffer]
			if !ok {
				log.Printf("no buffer for entity: %d\n", e)
				continue
			}
			b.SetActive()
			curBuf = s.Buffer
			drawBuffer = b
		}
		if curProgram != s.Program {
			p, ok := g.programs[s.Program]
			if !ok {
				log.Printf("no program for entity: %d\n", e)
				continue
			}
			curProgram = s.Program
			SetActiveProgram(p)
			program = p
			err := AssignUniformMat4(program, "projView", projView)
			if err != nil {
				log.Println(err)
			}
		}
		if curTexture != s.Textures {
			tex, ok := g.textures[s.Textures]
			if ok {
				curTexture = s.Textures
				if s.CurTexture == -1 {
					for i, t := range tex {
						SetActiveTex(t, i)
					}
				} else {
					SetActiveTex(tex[s.CurTexture], 0)
				}
			}
		}
		t := w.GetComponent(e, yecs.TransformComponent).(yecs.Transform)
		err := AssignUniformMat4(program, "world", t.Transform)
		if err != nil {
			log.Println(err)
		}
		if s.AssignUniforms != nil {
			err := s.AssignUniforms(e, w, &MainCam, program)
			if err != nil {
				log.Println(err)
			}
		}
		r.SetupRenderState()
		drawBuffer.DrawBuffer()
	}
	//elapsed := time.Since(now)
	//	fmt.Println(elapsed)
	g.Window.GLSwap()
}
