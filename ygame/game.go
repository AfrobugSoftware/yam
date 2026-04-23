package ygame

import (
	"log/slog"
	"os"
	"time"
	"yam/yecs"
	"yam/ygl"

	"github.com/ebitengine/oto/v3"
	"github.com/veandco/go-sdl2/sdl"
)

var (
	MS_PER_FRAME = 16 * time.Millisecond
)

type Game struct {
	World      *yecs.World
	Running    bool
	Ticks      uint64
	NeedsReset bool
	ShowGrid   bool
	IGrid      *Grid
	DoReset    func()
	OnExit     func() bool
	logFile    *os.File
	Log        *slog.Logger
	Gl3        *ygl.Gl3
	Audio      *yecs.AudioSystem
	Input      *yecs.InputSystem
}

var gGame *Game

func NewGame(title string, width, height int32) (*Game, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, err
	}
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLEBUFFERS, 1) // enable multisampling
	sdl.GLSetAttribute(sdl.GL_MULTISAMPLESAMPLES, 4) // 4x MSAA (2, 4, 8, 16)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 3)
	sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
	sdl.GLSetAttribute(sdl.GL_DOUBLEBUFFER, 1)
	sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 24)
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height,
		sdl.WINDOW_OPENGL|sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile("yam.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	handler := slog.NewJSONHandler(file, nil)
	logger := slog.New(handler)

	gl3, err := ygl.NewYGL(window, int(width), int(height))
	if err != nil {
		return nil, err
	}

	gGame = &Game{
		logFile: file,
		Log:     logger,
		Ticks:   sdl.GetTicks64(),
		Gl3:     gl3,
		World:   yecs.NewWorld(),
		Audio:   yecs.NewAudioSystem(yecs.STEREO, 44000, oto.FormatFloat32LE),
		Input: &yecs.InputSystem{
			ShowCursor:   1,
			ScreenWidth:  width,
			ScreenHeight: height,
		},
		ShowGrid: true,
		//IGrid:    NewGrid(),
	}
	return gGame, nil
}

func GetGame() *Game {
	return gGame
}

func (g *Game) Update(dt float64) {
	g.World.Update(dt)
}

func (g *Game) ProcessInput() {
	g.Input.PrepareToUpdate()
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.GetType() {
		case sdl.QUIT:
			g.Running = false
			if g.OnExit != nil {
				g.Running = g.OnExit()
			}
		case sdl.MOUSEWHEEL:
			w := event.(*sdl.MouseWheelEvent)
			g.Input.ProcessWheel(w)
		}
		state := sdl.GetKeyboardState()
		if state != nil {
			if state[sdl.SCANCODE_ESCAPE] != 0 {
				g.Running = false
			}
			copy(g.Input.CurKeyState, state)
		}
		g.Input.UpdateMouse()
	}
	g.Input.UpdateInput(g.World)
}

func (g *Game) Draw() {
	// if g.ShowGrid {
	// 	g.IGrid.Draw(g.World)
	// }
	g.Gl3.DrawSpatial(g.World)

}

func (g *Game) Run() {
	defer g.Quit()
	var dt time.Duration
	g.Running = true
	lastTime := time.Now()
	for g.Running {
		now := time.Now()
		dt = now.Sub(lastTime)
		frameTime := dt.Seconds()
		if frameTime > 0.05 {
			frameTime = 0.05
		}
		g.ProcessInput()
		g.Update(frameTime)
		g.Draw()
		lastTime = now
	}
}

func (g *Game) Quit() {
	g.World.Shutdown()
	if g.Gl3 != nil {
		g.Gl3.ShutDownGL()
	}
	if g.logFile != nil {
		g.logFile.Close()
	}
}
