package yam

import (
	"math/rand"
	"yam/y3d"
	"yam/yecs"
	"yam/ygame"
	"yam/ygl"

	gl "github.com/chsc/gogl/gl33"
)

const (
	height = 720
	width  = 1280
)

func CreateObject(w *yecs.World, transform yecs.Transform) {
	ent := w.NewEntity()
	renderState := yecs.RenderState{}
	renderState.States = append(renderState.States, yecs.DepthState{
		Enable:    true,
		DepthFunc: gl.LESS,
	}, yecs.BlendState{
		SrcFactor: gl.DST_ALPHA,
		DstFactor: gl.ONE_MINUS_DST_ALPHA,
	})
	sprite := yecs.Sprite{
		Buffer:   "sprite",
		Textures: "sprite",
		Program:  "sprite",
	}
	w.AddComponent(ent, yecs.SpriteComponent, sprite)
	w.AddComponent(ent, yecs.TransformComponent, transform)
	w.AddComponent(ent, yecs.RenderStateComponent, renderState)
}

func CreateResources(g *ygame.Game) {
	g.Gl3.AddSprite("sprite")
	err := g.Gl3.AddProgramSource("sprite", ygl.SpriteVert, ygl.SpriteFrag)
	if err != nil {
		panic(err)
	}
}

func CreateCamera(w *yecs.World) {
	ent := w.NewEntity()
	camera := yecs.Camera{
		Pos:             y3d.ZEROV,
		Up:              y3d.UNIT_Y,
		LookAt:          y3d.UNIT_Z,
		NeedCalculation: true,
		Speed:           5,
		Right:           width,
		Left:            0,
		Top:             0,
		Bottom:          height,
		Near:            10,
		Far:             10000,
		CamType:         yecs.CAM_TYPE_PESPECTIVE,
	}
	w.AddComponent(ent, yecs.CameraComponent, camera)
}

func randRange(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func CreateScene(w *yecs.World) {
	for range 100 {
		x := randRange(0, 1000)
		y := randRange(0, 1000)
		z := randRange(100, 1000)
		transform := yecs.Transform{
			Position: y3d.Vec3{X: x, Y: y, Z: z},
			Scale:    y3d.Vec3{X: 20, Y: 20, Z: 1},
		}
		CreateObject(w, transform)
	}
	CreateCamera(w)
}

func TestGame() {
	g, err := ygame.NewGame("Test scene", width, height)
	if err != nil {
		panic(err)
	}
	CreateResources(g)
	CreateScene(g.World)

	g.Run()
}
