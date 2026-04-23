package ygl

import (
	"fmt"
	"log"
	"sync"
	"unsafe"
	"yam/y3d"
	"yam/yecs"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

type Gl3 struct {
	Context      sdl.GLContext
	Window       *sdl.Window
	ClearColor   y3d.Vec4
	PixelDepth   uint8
	DoubleBuffer bool
	MinorVersion int
	MajorVersion int
	Height       int32
	Width        int32

	mu       sync.Mutex
	buffers  map[string]VertBuffer
	programs map[string]uint32
	textures map[string][]uint32
}

func NewYGL(window *sdl.Window, width, height int) (*Gl3, error) {
	g := &Gl3{
		Window:   window,
		buffers:  make(map[string]VertBuffer),
		programs: make(map[string]uint32),
		textures: make(map[string][]uint32),
		Height:   int32(height),
		Width:    int32(width),
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
	texs := make([]uint32, 0, len(filePath))
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

func (g *Gl3) AddPrograms(name string, filePath []string, types []uint32) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if len(filePath) != len(types) {
		return fmt.Errorf("invalid file to types")
	}
	shaders := []uint32{}
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
	p, err := CreateProgram([]uint32{v, f})
	if err != nil {
		return err
	}
	g.programs[name] = p
	return nil
}

func (g *Gl3) AddSprite(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	buf := unsafe.Slice((*byte)(unsafe.Pointer(&SpriteData[0])), len(SpriteData)*int(unsafe.Sizeof(float32(0))))
	vbuf := CreateVextexBuffer(buf, SpriteIndices, SpriteFormat[:])
	g.buffers[name] = vbuf
}

func (g *Gl3) AddVertexBuffer(name string, data []byte, indx []uint16, formats []DataFormat) {
	g.mu.Lock()
	defer g.mu.Unlock()

	vbuf := CreateVextexBuffer(data, indx, formats)
	g.buffers[name] = vbuf
}

func (g *Gl3) DrawSpatial(w *yecs.World) {
	//now := time.Now()
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	camera := w.Query([]yecs.ComponentId{yecs.CameraComponent})
	sprites := w.Query([]yecs.ComponentId{yecs.SpatialComponent, yecs.TransformComponent, yecs.RenderStateComponent})
	//animatedSpatials = w.Query([]yecs.ComponentId{yecs.AnimatedSpatialCompnent, yecs.TransformComponent, yecs.RenderStateComponent})
	if len(camera) == 0 {
		log.Println("no camera attached to scene")
		return
	}
	MainCam := w.GetComponent(camera[0], yecs.CameraComponent).(yecs.Camera)
	view := MainCam.GetViewTransformation()
	proj := MainCam.GetProjectionTransformation()
	projView := proj.Mul(view)
	var curBuf, curProgram, curTexture string
	var program uint32
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
						SetActiveTex(t, uint32(i))
					}
				} else {
					SetActiveTex(tex[s.CurTexture], 0)
				}
			}
		}
		t := w.GetComponent(e, yecs.TransformComponent).(yecs.Transform)
		err := AssignUniformMat4(program, "world", t.Local)
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
