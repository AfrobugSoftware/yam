package ygame

import (
	"log"
	"log/slog"
	"os"
	"yam/yecs"
	"yam/ygl"
	"yam/yrender"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type Game struct {
	World      *yecs.World
	Running    bool
	Renderer   *yrender.Renderer
	Ticks      uint64
	NeedsReset bool
	DoReset    func()
	logFile    *os.File
	Log        *slog.Logger
	GL         *ygl.Gl3
}

var gGame *Game

func NewGame(title string, width, height int32) (*Game, error) {
	r, err := yrender.CreateSDLRenderer(title, width, height)
	if err != nil {
		return nil, err
	}
	err = img.Init(img.INIT_PNG | img.INIT_JPG)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile("yam.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	handler := slog.NewJSONHandler(file, nil)
	logger := slog.New(handler)

	gGame = &Game{
		Renderer: r,
		logFile:  file,
		Log:      logger,
		Ticks:    sdl.GetTicks64(),
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
	g.Renderer.BeginRendering()

	g.Renderer.EndRendering()
}

func (g *Game) Run() {
	defer g.Quit()
	var dt float64
	g.Running = true
	for g.Running {
		//how do I wait for 16ms to pass ??

		dt = float64(sdl.GetTicks64()-g.Ticks) / 1000.0
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
	if g.Renderer != nil {
		g.Renderer.Close()
	}
	if g.logFile != nil {
		g.logFile.Close()
	}
}
