package yam

// import (
// 	"math"
// 	"yam/y3d"
// 	"yam/ygame"

// 	"github.com/veandco/go-sdl2/sdl"
// )

// const (
// 	height = 720
// 	width  = 1280
// )

// // actor types
// const (
// 	PADDLE = iota
// 	WALL
// 	BALL
// )

// type Paddle struct {
// 	ygame.Sprite
// 	Controls [2]int
// }

// type Wall struct {
// 	ygame.Sprite
// }

// type Ball struct {
// 	ygame.Sprite
// }

// func NewPaddle(pos y3d.Vec3, color [4]uint8, controls [2]int) *Paddle {
// 	p := &Paddle{}
// 	p.Width = 35
// 	p.Height = 100
// 	p.Pos = pos
// 	p.Color = color
// 	p.Controls = controls
// 	p.Components = append(p.Components, &ygame.MoveComponent{})
// 	return p
// }
// func (p *Paddle) GetType() int {
// 	return PADDLE
// }

// func (p *Paddle) ProcessInput() {
// 	g := ygame.GetGame()
// 	p.Vel.Y = 0
// 	speed := 250
// 	if g.KeyState != nil {
// 		if g.KeyState[p.Controls[0]] != 0 {
// 			p.Vel.Y -= float64(speed)
// 		}
// 		if g.KeyState[p.Controls[1]] != 0 {
// 			p.Vel.Y += float64(speed)
// 		}
// 	}
// }

// func (p *Paddle) Update(dt float64) {
// 	g := ygame.GetGame()
// 	oldPos := p.Pos
// 	p.Sprite.Update(dt)
// 	for i := range g.Actors {
// 		if g.Actors[i].GetType() == WALL {
// 			if p.Box.Overlaps(g.Actors[i].GetBox()) {
// 				p.Pos = oldPos
// 			}
// 		}
// 	}
// }

// func NewWall(pos y3d.Vec3, height int32, color [4]uint8) *Wall {
// 	g := ygame.GetGame()
// 	w := &Wall{}
// 	w.Pos = pos
// 	w.Color = color
// 	w.Width = g.Renderer.Width
// 	w.Height = int32(height)
// 	return w
// }

// func (w *Wall) GetType() int {
// 	return WALL
// }

// func (w *Wall) Update(_ float64) {
// 	w.GenerateAABB()
// }

// func NewBall(pos y3d.Vec3, color [4]uint8) *Ball {
// 	b := &Ball{}
// 	b.Pos = pos
// 	b.Vel = y3d.Vec3{X: -200.0, Y: 235.0, Z: 0}
// 	b.Width = 25
// 	b.Height = 25
// 	b.Color = color
// 	b.Components = append(b.Components, &ygame.MoveComponent{})
// 	return b
// }

// func (b *Ball) Update(dt float64) {
// 	g := ygame.GetGame()
// 	b.Sprite.Update(dt)
// 	for i := range g.Actors {
// 		switch g.Actors[i].GetType() {
// 		case WALL:
// 			w := g.Actors[i].GetBox()
// 			if b.Vel.Y < 0.0 && b.Pos.Y <= w.MaxY && b.Pos.Y >= w.MinY {
// 				b.Vel.Y *= -1
// 			} else if b.Vel.Y > 0.0 &&
// 				b.Pos.Y >= w.MinY && b.Pos.Y <= w.MaxY {
// 				b.Vel.Y *= -1
// 			}
// 		case PADDLE:
// 			p := g.Actors[i].(*Paddle)
// 			dist := math.Abs(p.Pos.Y - b.Pos.Y)
// 			if dist <= float64(p.Height)/2.0 && b.Vel.X < 0.0 &&
// 				b.Pos.X >= p.Box.MinX && b.Pos.X <= p.Box.MaxX {
// 				b.Vel.X *= -1
// 			} else if dist <= float64(p.Height)/2.0 && b.Vel.X > 0.0 &&
// 				b.Pos.X >= p.Box.MinX && b.Pos.X <= p.Box.MaxX {
// 				b.Vel.X *= -1
// 			}
// 		}
// 	}

// 	if b.Pos.X < 0.0 || b.Pos.X > float64(g.Renderer.Width) {
// 		g.NeedsReset = true
// 	}
// }

// func (b *Ball) GetType() int {
// 	return BALL
// }

// func initGame(g *ygame.Game) {
// 	p := NewPaddle(y3d.Vec3{
// 		X: 35,
// 		Y: (float64(g.Renderer.Height) / 2.0),
// 	},
// 		[4]uint8{255, 0, 0, 255},
// 		[2]int{sdl.SCANCODE_W, sdl.SCANCODE_S})

// 	p2 := NewPaddle(y3d.Vec3{
// 		X: float64(g.Renderer.Width) - 35,
// 		Y: (float64(g.Renderer.Height) / 2.0),
// 	},
// 		[4]uint8{0, 0, 255, 255},
// 		[2]int{sdl.SCANCODE_UP, sdl.SCANCODE_DOWN})

// 	w := NewWall(y3d.Vec3{
// 		X: float64(g.Renderer.Width) / 2,
// 		Y: 50 / 2,
// 	},
// 		50,
// 		[4]uint8{225, 178, 102, 255},
// 	)

// 	w2 := NewWall(y3d.Vec3{
// 		X: float64(g.Renderer.Width) / 2,
// 		Y: float64(g.Renderer.Height) - (50 / 2),
// 	},
// 		50,
// 		[4]uint8{225, 178, 102, 255},
// 	)

// 	b := NewBall(
// 		y3d.Vec3{
// 			X: float64(g.Renderer.Width) / 2,
// 			Y: float64(g.Renderer.Height) / 2,
// 		},
// 		[4]uint8{102, 0, 0, 255},
// 	)

// 	g.AddActor(p)
// 	g.AddActor(p2)
// 	g.AddActor(w)
// 	g.AddActor(w2)
// 	g.AddActor(b)
// }

// func NewPingPongGame() {
// 	g, err := ygame.NewGame("ping pong game", width, height)
// 	if err != nil {
// 		panic(err)
// 	}
// 	initGame(g)
// 	g.DoReset = func() {
// 		g.Actors = nil
// 		initGame(g)
// 	}
// 	g.Run()
// }
