package yam

import (
	"fmt"
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

func CreateObject(w *yecs.World, transform yecs.Transform, curText int) {
	ent := w.NewEntity()
	renderState := yecs.RenderState{}
	renderState.States = append(renderState.States, yecs.DepthState{
		Enable:    true,
		DepthFunc: gl.LESS,
	}, yecs.BlendState{
		Enable:    true,
		SrcFactor: gl.SRC_ALPHA,
		DstFactor: gl.ONE_MINUS_SRC_ALPHA,
	})
	sprite := yecs.Sprite{
		Buffer:     "sprite",
		Textures:   "sprite",
		Program:    "sprite",
		CurTexture: curText,
	}
	move := yecs.Move{
		AnglularSpeed: 5,
		ForwardSpeed:  20,
	}
	w.AddComponent(ent, yecs.SpriteComponent, sprite)
	w.AddComponent(ent, yecs.TransformComponent, transform)
	w.AddComponent(ent, yecs.RenderStateComponent, renderState)
	w.AddComponent(ent, yecs.MoveComponent, move)
}

func CreateResources(g *ygame.Game) {
	g.Gl3.AddSprite("sprite")
	err := g.Gl3.AddProgramSource("sprite", ygl.SpriteVert, ygl.SpriteFrag)
	if err != nil {
		panic(err)
	}
	err = g.Gl3.AddTextures([]string{
		"assets/orange.png",
		"assets/pear.png",
		"assets/oats.png",
	}, "sprite")
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
		Speed:           20,
		Right:           1,
		Left:            -1,
		Top:             0.75,
		Bottom:          -0.75,
		Near:            0.1,
		Far:             1000,
		CamType:         yecs.CAM_TYPE_PERSPECTIVE,
	}
	w.AddComponent(ent, yecs.CameraComponent, camera)
}

func randRange(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func CreateScene(w *yecs.World) {
	quat := y3d.FromAngleAxis(y3d.UNIT_X, 0)
	for i := range 1000 {
		x := randRange(-1000, 1000)
		y := randRange(-1000, 1000)
		z := randRange(-10, -1000)
		transform := yecs.Transform{
			Name:            fmt.Sprintf("transform: %d", i),
			Position:        y3d.Vec3{X: x, Y: y, Z: z},
			Scale:           y3d.Vec3{X: 64, Y: 64, Z: 1},
			Orientation:     quat,
			NeedCalculation: true,
		}
		CreateObject(w, transform, i%3)
	}
	CreateCamera(w)
	w.AddSystem(&yecs.StateSystem{})
	w.AddSystem(&yecs.MoveSystem2D{})
	w.AddSystem(&yecs.AudioSystem{})
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
