package ygl

import (
	"yam/yecs"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type SpriteBatch struct {
	numSpriteX         float32
	numSpriteY         float32
	windowWidth        float32
	windowHeight       float32
	spriteAspectRatio  float32
	texUSize           float32
	texVSize           float32
	imageWidth         float32
	imageHeight        float32
	windowAspectRatio  float32
	ndcPixelX          float32
	ndcPixelY          float32
	tileHeightInPixels float32
	tileWidthInPixels  float32
	tileHeightNDC      float32
	tileWidthNDC       float32
	spreedSheetTex     uint32
	shader             uint32
	ubo                uint32
	quadArray          *QuadArray
}

func CreateSpriteBatch(filename string,
	numSpriteX, numSpriteY, windowWidth,
	windowHeight, imageWidth, imageHeight float32) *SpriteBatch {
	sp := &SpriteBatch{}
	vs, err := CreateShaderFromFile("assets/shaders/sprite.vert", gl.VERTEX_SHADER)
	if err != nil {
		panic(err)
	}
	fs, err := CreateShaderFromFile("assets/shaders/sprite.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		panic(err)
	}
	p, err := CreateProgram([]uint32{vs, fs})
	if err != nil {
		panic(err)
	}
	tex, err := CreateTex2D(filename, gl.NEAREST, gl.NEAREST, false)
	if err != nil {
		panic(err)
	}

	sp.shader = p
	sp.spreedSheetTex = tex
	sp.numSpriteX = numSpriteX
	sp.numSpriteY = numSpriteY
	sp.windowWidth = windowWidth
	sp.windowHeight = windowHeight
	sp.imageWidth = imageWidth
	sp.imageHeight = imageHeight
	sp.windowAspectRatio = windowHeight / windowWidth
	sp.quadArray = CreateQuadArray()
	sp.CalculateSpriteInfo()
	sp.quadArray.SetupUBO(p)
	return sp
}

func (sp *SpriteBatch) CalculateSpriteInfo() {
	spriteWidth := sp.imageWidth / sp.numSpriteX
	spriteHeigth := sp.imageHeight / sp.numSpriteY
	sp.spriteAspectRatio = spriteHeigth / spriteWidth
	sp.texUSize = 1.0 / sp.numSpriteX
	sp.texVSize = 1.0 / sp.numSpriteY

	imageWidthToWindowWidthRatio := sp.imageWidth / sp.windowWidth
	imageHeighToWindowHeightRatio := sp.imageHeight / sp.windowHeight
	if imageWidthToWindowWidthRatio < imageHeighToWindowHeightRatio {
		sp.tileHeightInPixels = sp.windowHeight / sp.numSpriteY
		sp.tileWidthInPixels = sp.tileHeightInPixels / sp.spriteAspectRatio
	} else {
		sp.tileWidthInPixels = sp.windowWidth / sp.numSpriteX
		sp.tileHeightInPixels = sp.tileWidthInPixels / sp.spriteAspectRatio
	}

	sp.ndcPixelX = 2.0 / sp.windowWidth
	sp.ndcPixelY = 2.0 / sp.windowHeight
	sp.tileWidthNDC = sp.ndcPixelX * sp.tileWidthInPixels
	sp.tileHeightNDC = sp.ndcPixelY * sp.tileHeightInPixels
}

func (sp *SpriteBatch) ScreenPosToNDC(x, y float32) (nx, ny float32) {
	nx = (2.0*x)/sp.windowWidth - 1.0
	ny = (2.0*y)/sp.windowHeight - 1.0
	return
}

func (sp *SpriteBatch) Draw(w *yecs.World, entites []yecs.EntityId) {
	gl.UseProgram(sp.shader)
	var ndcX, ndcY float32
	for i, e := range entites {
		s := w.GetComponent(e, yecs.SpriteComponent).(yecs.Sprite)
		ndcX, ndcY = sp.ScreenPosToNDC(float32(s.Pos[0]), float32(s.Pos[1]))
		tileWidthNDC := sp.ndcPixelX * float32(s.Width)
		tileHeightNDC := tileWidthNDC / sp.spriteAspectRatio
		UBase := sp.texUSize * float32(s.Col)
		VBase := sp.texVSize * float32(s.Row)
		sp.quadArray.SetQuad(i, ndcX, ndcY, tileWidthNDC, tileHeightNDC, UBase, VBase, sp.texUSize, sp.texVSize)
	}
	sp.quadArray.Update()
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, sp.spreedSheetTex)
	sp.quadArray.Draw(len(entites))
	gl.UseProgram(0)
}
