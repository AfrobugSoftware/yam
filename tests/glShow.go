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

func CreateObject(w *yecs.World, vbo ygl.VertBuffer, transform yecs.Transform, curText int, surface *yecs.MaterialSurface) {
	ent := w.NewEntity()
	renderState := yecs.RenderState{}
	renderState.States = append(renderState.States, yecs.DepthState{
		Enable:    true,
		DepthFunc: gl.LESS,
	}, yecs.BlendState{
		Enable:    false,
		SrcFactor: gl.SRC_ALPHA,
		DstFactor: gl.ONE_MINUS_SRC_ALPHA,
	},
		yecs.FaceState{
			Enable:    true,
			CullFace:  gl.BACK,
			FrontFace: gl.CCW,
		},
	)
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
		AnglularSpeed: 15 * math.Pi,
		ForwardSpeed:  0,
	}
	h := yecs.Hierarchy{
		Parent: yecs.NullEntity,
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
	w.AddComponent(ent, yecs.MaterialSurfaceComponent, *surface)
	w.AddComponent(ent, yecs.HierarchyComponent, h)
}

func CreateCamera(w *yecs.World) {
	ent := w.NewEntity()
	camera := yecs.Camera{
		Pos:    y3d.ZEROV,
		Up:     yecs.UP,
		LookAt: yecs.FORWARD,
		Speed:  20,
		// Right:   1,
		// Left:    -1,
		// Top:     0.75,
		// Bottom:  -0.75,
		Width:   width,
		Height:  height,
		Fov:     90,
		Near:    0.1,
		Far:     10000,
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
	tex, err := ygl.CreateTex2D("assets/earth.jpg", gl.LINEAR_MIPMAP_LINEAR, gl.LINEAR, true)
	if err != nil {
		panic(err)
	}
	surface := yecs.MaterialSurface{
		Diffuse: tex,
	}
	for i := range 100 {
		x := randRange(-1000, 1000)
		y := randRange(-1000, 1000)
		z := randRange(10, 1000)
		transform := yecs.Transform{
			Position: y3d.Vec3{X: x, Y: y, Z: z},
			Scale:    y3d.Vec3{X: 64, Y: 64, Z: 64},
			Rotation: y3d.IdenQuat(),
		}
		CreateObject(w, vbo, transform, i%3, &surface)
	}
	CreateCamera(w)
}

func CreateSystems(w *yecs.World) {
	w.AddSystem(ygame.GetGame().Audio)
	w.AddSystem(ygame.GetGame().Input)
	w.AddSystem(&yecs.CameraSystem{})
	w.InitSystems()
}

func TestLoadAssets() {
	g, err := ygame.NewGame("Test scene", width, height)
	if err != nil {
		panic(err)
	}
	_, err = ygl.LoadAsset("assets/gltf/testgltf.gltf", g.World)
	if err != nil {
		panic(err)
	}
}

func TestGame() {
	g, err := ygame.NewGame("Test scene", width, height)
	if err != nil {
		panic(err)
	}
	CreateSystems(g.World)
	CreateLight(g.World, 3)
	CreateScene(g.World)

	_, err = ygl.LoadAsset("assets/gltf/testgltf.gltf", g.World)
	if err != nil {
		panic(err)
	}

	g.Run()
}
