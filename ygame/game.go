package ygame

import (
	"log"
	"log/slog"
	"os"
	"yam/yrender"

	"github.com/veandco/go-sdl2/sdl"
)

type Game struct {
	Actors     []Actor
	DeadActors []Actor
	Running    bool
	ClearColor [4]uint8
	//input state
	KeyState []uint8
	MouseX   float32
	MouseY   float32
	Renderer *yrender.Renderer

	Ticks   uint64
	logFile *os.File
	Log     *slog.Logger
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
		Renderer:   r,
		logFile:    file,
		Log:        logger,
		ClearColor: [4]uint8{255, 255, 255, 255},
		Ticks:      sdl.GetTicks64(),
	}
	return gGame, nil
}

func GetGame() *Game {
	return gGame
}

func (g *Game) Update(dt float64) {
	for _, a := range g.Actors {
		a.Update(dt)
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
		}
		g.KeyState = make([]uint8, 0, len(state))
		copy(g.KeyState, state)
	}
}

func (g *Game) Draw() {
	g.Renderer.Clear()
	//rest of the drawing goes here
	for _, a := range g.Actors {
		a.Draw()
	}
	g.Renderer.Swap()
}

func (g *Game) Run() {
	defer g.Quit()
	var dt float64
	g.Running = true
	g.Renderer.ClearBackgroundColor(
		g.ClearColor[0],
		g.ClearColor[1],
		g.ClearColor[2],
		g.ClearColor[3],
	)
	for g.Running {
		//wait for 16ms to pass
		for int64(g.Ticks-sdl.GetTicks64()) <= 0 {
		}

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
