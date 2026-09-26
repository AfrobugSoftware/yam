package ygame

import (
	"time"
	"yam/ycore"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	MS_PER_FRAME = 16 * time.Millisecond
)

type Engine struct {
	App           Application
	RenderManager *ycore.RenderManager
	InputManager  *ycore.InputManager
	AudioManager  *ycore.AudioManager
	NetManager    *ycore.NetManager
}

var gEngine *Engine

func NewGame(title string, width, height int32) (*Engine, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, err
	}
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1) // enable multisampling
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4) // 4x MSAA (2, 4, 8, 16)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 4)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 24)
	sdl.GLSetSwapInterval(1)
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height,
		sdl.WINDOW_OPENGL|sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	gEngine = &Engine{}
	gEngine.RenderManager = ycore.NewRenderManager(window, int(width), int(height))
	gEngine.InputManager = ycore.NewInputManager(int(width), int(height))
	gEngine.AudioManager = ycore.NewAudioManager()
	return gEngine, nil
}

func GetGame() *Engine {
	return gEngine
}

func (g *Engine) SetApplication(app Application) {
	g.App = app
	g.App.Startup(g)
}

func (g *Engine) Update(dt float64) {
	if g.App != nil {
		g.App.Update(dt)
	}
}

func (g *Engine) Draw() {
	g.RenderManager.BeginRender()
	if g.App != nil {
		g.App.Draw()
	}
	g.RenderManager.Render() //how to I allow the frontend render
	g.RenderManager.EndRender()
}

func (g *Engine) Run() {
	defer g.Quit()
	var dt time.Duration
	lastTime := time.Now()
	for g.InputManager.ProcessInput() {
		now := time.Now()
		dt = now.Sub(lastTime)
		frameTime := dt.Seconds()
		if frameTime > 0.05 {
			frameTime = 0.05
		}
		g.Update(frameTime)
		g.Draw()
		lastTime = now
	}
}

func (g *Engine) Quit() {
	if g.App != nil {
		g.App.Shutdown()
	}
	g.RenderManager.Destroy()
	sdl.Quit()
}
