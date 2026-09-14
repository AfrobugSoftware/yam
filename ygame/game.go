package ygame

import (
	"os"
	"time"
	"yam/ymanager"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	MS_PER_FRAME = 16 * time.Millisecond
)

type Game struct {
	Running       bool
	Ticks         uint64
	NeedsReset    bool
	ShowGrid      bool
	DoReset       func()
	OnExit        func() bool
	logFile       *os.File
	App           Application
	RenderManager *ymanager.RenderManager
	InputManager  *ymanager.InputManager
}

var gGame *Game

func NewGame(title string, width, height int32) (*Game, error) {
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
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height,
		sdl.WINDOW_OPENGL|sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	gGame = &Game{
		Ticks: sdl.GetTicks64(),
	}
	gGame.RenderManager = ymanager.NewRenderManager(window, int(width), int(height))
	gGame.InputManager = ymanager.NewInputManager()

	return gGame, nil
}

func GetGame() *Game {
	return gGame
}

func (g *Game) SetApplication(app Application) {
	g.App = app
	g.App.Startup()
}

func (g *Game) Update(dt float64) {
	if g.App != nil {
		g.App.Update(time.Now())
	}
}

func (g *Game) Draw() {
	if g.App != nil {
		g.App.Draw()
	}
	g.RenderManager.Render() //how to I allow the frontend render
}

func (g *Game) Run() {
	defer g.Quit()
	var dt time.Duration
	g.Running = true
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

func (g *Game) Quit() {
	if g.App != nil {
		g.App.Shutdown()
	}
	g.RenderManager.Destroy()
	sdl.Quit()
}
