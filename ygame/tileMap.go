package ygame

import "yam/yecs"

type TileMap struct {
	Level       []yecs.EntityId
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
	//how to laod the map

	return &TileMap{
		Level:       []yecs.EntityId{},
		LevelWidth:  width,
		LevelHeight: height,
		TileWidth:   tileWidth,
		TileHeight:  tileHeight,
		Camara:      camara,
	}
}

func (t *TileMap) GetTile(x, y int) yecs.EntityId {
	if x >= 0 && x < t.LevelWidth && y >= 0 && y < t.LevelHeight {
		return t.Level[y*t.LevelWidth+x]
	} else {
		return NULLCHARACTER
	}
}

func (t *TileMap) SetTile(x, y int, r yecs.EntityId) {
	if x >= 0 && x < t.LevelWidth && y >= 0 && y < t.LevelHeight {
		t.Level[y*t.LevelWidth+x] = r
	}
}
