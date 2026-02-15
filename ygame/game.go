package ygame

import (
	"log"
	"log/slog"
	"os"
	"yam/yrender"

	"github.com/veandco/go-sdl2/sdl"
)

type Game struct {
	Actors        []Actor
	DeadActors    []Actor
	SpawnedActors []Actor
	Running       bool
	//input state
	KeyState []uint8
	MouseX   float32
	MouseY   float32
	Renderer *yrender.Renderer

	Ticks      uint64
	NeedsReset bool
	DoReset    func()
	logFile    *os.File
	Log        *slog.Logger
}

var gGame *Game

func NewGame(title string, width, height int32) (*Game, error) {
	r, err := yrender.CreateSDLRenderer(title, width, height)
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
func (g *Game) AddActor(a Actor) {
	g.Actors = append(g.Actors, a)
}

func (g *Game) AddSpawnedActor(a Actor) {
	g.Actors = append(g.SpawnedActors, a)
}

func GetGame() *Game {
	return gGame
}

func (g *Game) Update(dt float64) {
	for i := range g.Actors {
		g.Actors[i].Update(dt)
	}
	if g.NeedsReset {
		if g.DoReset != nil {
			g.DoReset()
			g.NeedsReset = false
		}
	}
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
			if g.KeyState == nil {
				g.KeyState = make([]uint8, len(state))
			}
			copy(g.KeyState, state)
		}
	}
	for i := range g.Actors {
		g.Actors[i].ProcessInput()
	}
}

func (g *Game) Draw() {
	//rest of the drawing goes here
	g.Renderer.BeginRendering()
	for i := range g.Actors {
		g.Actors[i].Draw(g.Renderer.Renderer)
	}
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
