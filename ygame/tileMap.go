package ygame

import (
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

type TileMap struct {
	Level       []rune
	LevelWidth  int
	LevelHeight int
	TileWidth   int
	TileHeight  int
	Camara      *Camara
}

const (
	NULLCHARACTER = ' '
)

func NewTileMap(level string, width, height, tileWidth, tileHeight int, camara *Camara) *TileMap {
	return &TileMap{
		Level:       []rune(level),
		LevelWidth:  width,
		LevelHeight: height,
		TileWidth:   tileWidth,
		TileHeight:  tileHeight,
		Camara:      camara,
	}
}

func (t *TileMap) Draw(renderer *sdl.Renderer) {
	g := GetGame()
	visableTilesX := g.Renderer.Width / int32(t.TileWidth)
	visableTilesY := g.Renderer.Height / int32(t.TileHeight)

	offsetX := t.Camara.Pos.X - float64(visableTilesX)/2.0
	offsetY := t.Camara.Pos.Y - float64(visableTilesY)/2.0
	if offsetX < 0 {
		offsetX = 0
	}
	if offsetY < 0 {
		offsetY = 0
	}
	if offsetX > (float64(t.LevelWidth - int(visableTilesX))) {
		offsetX = float64(t.LevelWidth - int(visableTilesX))
	}
	if offsetY > (float64(t.LevelHeight - int(visableTilesY))) {
		offsetY = float64(t.LevelHeight - int(visableTilesY))
	}
	_, fx := math.Modf(offsetX)
	_, fy := math.Modf(offsetY)
	tileOffsetX := fx * float64(t.TileWidth)
	tileOffsetY := fy * float64(t.TileHeight)
	for x := 0; x < int(visableTilesX); x++ {
		for y := 0; y < int(visableTilesY); y++ {
			tileId := t.GetTile(x+int(offsetX), y+int(offsetY))
			rect := sdl.Rect{
				X: int32(float64(x)*float64(t.TileWidth) - tileOffsetX),
				Y: int32(float64(y)*float64(t.TileHeight) - tileOffsetY),
				W: int32(t.TileWidth),
				H: int32(t.TileHeight),
			}
			switch tileId {
			case '.':
				renderer.SetDrawColor(0, 255, 255, 255)
				renderer.FillRect(&rect)
			case '#':
				renderer.SetDrawColor(255, 0, 0, 255)
				renderer.FillRect(&rect)
			default:
			}
		}
	}
}

func (t *TileMap) GetTile(x, y int) rune {
	if x >= 0 && x < t.LevelWidth && y >= 0 && y < t.LevelHeight {
		return t.Level[y*t.LevelWidth+x]
	} else {
		return NULLCHARACTER
	}
}

func (t *TileMap) SetTile(x, y int, r rune) {
	if x >= 0 && x < t.LevelWidth && y >= 0 && y < t.LevelHeight {
		t.Level[y*t.LevelWidth+x] = r
	}
}
