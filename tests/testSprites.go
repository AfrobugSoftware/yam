package yam

import (
	"math"
	"math/rand"
	"yam/yecs"
	"yam/ygame"
)

func randomClusteredPixel(screenW, screenH int, spread float64) (int, int) {
	centerX := float64(screenW) / 2
	centerY := float64(screenH) / 2

	x := centerX + spread*rand.NormFloat64()
	y := centerY + spread*rand.NormFloat64()

	// Clamp to screen bounds
	x = math.Max(0, math.Min(float64(screenW-1), x))
	y = math.Max(0, math.Min(float64(screenH-1), y))

	return int(x), int(y)
}

func CreateSprites(count int, w *yecs.World) {
	for range count {
		e := w.NewEntity()
		x, y := randomClusteredPixel(width, height, 200.0)
		s := yecs.Sprite{}
		s.Width = 128
		s.Pos[0] = x
		s.Pos[1] = y
		s.Col = rand.Int() % 2
		s.Row = rand.Int() % 2

		w.AddComponent(e, yecs.SpriteComponent, s)
	}

}

func TestSprites() {
	g, err := ygame.NewGame("Test sprites", width, height)
	CreateSprites(500, g.World)
	if err != nil {
		panic(err)
	}
	g.Run()
}
