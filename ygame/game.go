package ygame

import (
	"log/slog"
	"os"
	"yam/yecs"
	"yam/ygl"

	"github.com/ebitengine/oto/v3"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type Game struct {
	World      *yecs.World
	Running    bool
	Ticks      uint64
	NeedsReset bool
	DoReset    func()
	OnExit     func() bool
	logFile    *os.File
	Log        *slog.Logger
	Gl3        *ygl.Gl3
	Audio      *yecs.AudioSystem
}

var gGame *Game

func NewGame(title string, width, height int32) (*Game, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, err
	}
	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height,
		sdl.WINDOW_OPENGL|sdl.WINDOW_ALLOW_HIGHDPI|sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}

	err = img.Init(img.INIT_PNG | img.INIT_JPG)
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
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.GetType() {
		case sdl.QUIT:
			g.Running = false
			if g.OnExit != nil {
				g.Running = g.OnExit()
			}
		}
		state := sdl.GetKeyboardState()
		if state != nil {
			if state[sdl.SCANCODE_ESCAPE] != 0 {
				g.Running = false
			}
			g.World.ProcessInput(state)
		}
	}
}

func (g *Game) Draw() {
	g.Gl3.DrawSprites(g.World)

}

func (g *Game) Run() {
	defer g.Quit()
	var dt float64
	g.Running = true
	for g.Running {
		//how do I wait for 16ms to pass ??
		dt = float64(sdl.GetTicks64()-g.Ticks) * 0.001
		g.Ticks = sdl.GetTicks64()
		if dt > 0.05 {
			dt = 0.05
		}
		g.ProcessInput()
		g.Update(dt)
		g.Draw()
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
