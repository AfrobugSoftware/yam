package yam

import (
	"math"
	"math/rand"
	"yam/y3d"
	"yam/yecs"
	"yam/ygame"
	"yam/ygl"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	height = 720
	width  = 1280
)

func CreateObject(w *yecs.World, vbo ygl.VertBuffer, transform yecs.Transform, curText int) {
	ent := w.NewEntity()
	renderState := yecs.RenderState{}
	renderState.States = append(renderState.States, yecs.DepthState{
		Enable:    true,
		DepthFunc: gl.LESS,
	}, yecs.BlendState{
		Enable:    true,
		SrcFactor: gl.SRC_ALPHA,
		DstFactor: gl.ONE_MINUS_SRC_ALPHA,
	},
		yecs.FaceState{
			Enable:    true,
			CullFace:  gl.BACK,
			FrontFace: gl.CCW,
		})
	sprite := yecs.Spatial{
		Buf:            vbo.Buf,
		Indx:           vbo.Indx,
		VertArray:      vbo.VertArray,
		IndxCount:      vbo.IndxCount,
		Program:        0,
		CurTexture:     curText,
		AssignUniforms: AddUniforms,
	}
	move := yecs.Move{
		AnglularSpeed: 5 * math.Pi,
		ForwardSpeed:  20,
	}

	// aabb := ygl.MakeAABBForSprite(ygl.SpriteData[:], ygl.SpriteFormat[0])
	// box := yecs.Box{
	// 	Local: aabb,
	// 	World: aabb,
	// }
	//w.AddComponent(ent, yecs.BoxComponent, box)
	w.AddComponent(ent, yecs.SpatialComponent, sprite)
	w.AddComponent(ent, yecs.TransformComponent, transform)
	w.AddComponent(ent, yecs.RenderStateComponent, renderState)
	w.AddComponent(ent, yecs.MoveComponent, move)
}

func CreateCamera(w *yecs.World) {
	ent := w.NewEntity()
	camera := yecs.Camera{
		Pos:     y3d.ZEROV,
		Up:      yecs.UP,
		LookAt:  yecs.FORWARD,
		Speed:   20,
		Right:   1,
		Left:    -1,
		Top:     0.75,
		Bottom:  -0.75,
		Near:    0.1,
		Far:     1000,
		CamType: yecs.CAM_TYPE_PERSPECTIVE,
	}
	camera.Recalulate()
	w.AddComponent(ent, yecs.CameraComponent, camera)
}

func randRange(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func CreateScene(w *yecs.World) {
	buffer, indices, format := ygl.CreateSphere(56, 28, 1.0)
	vbo := ygl.CreateVextexBuffer(buffer, indices, format)

	for i := range 1000 {
		x := randRange(-1000, 1000)
		y := randRange(-1000, 1000)
		z := randRange(-10, -1000)
		transform := yecs.Transform{
			Position: y3d.Vec3{X: x, Y: y, Z: z},
			Scale:    y3d.Vec3{X: 64, Y: 64, Z: 1},
			Rotation: y3d.IdenQuat(),
		}
		CreateObject(w, vbo, transform, i%3)
	}
	CreateCamera(w)
}

func CreateSystems(w *yecs.World) {
	w.AddSystem(ygame.GetGame().Audio)
	w.AddSystem(ygame.GetGame().Input)
	w.AddSystem(&yecs.CameraSystem{})
	w.InitSystems()
}

func TestGame() {
	g, err := ygame.NewGame("Test scene", width, height)
	if err != nil {
		panic(err)
	}
	CreateSystems(g.World)
	CreateLight(g.World)
	CreateScene(g.World)

	g.Run()
}
