package ygame

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
