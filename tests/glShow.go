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

func CreateObject(w *yecs.World, me yecs.MeshEntry, transform yecs.Transform, curText int, surface *yecs.MaterialSurface) yecs.EntityId {
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
	move := yecs.Move{
		AnglularSpeed: 10 * math.Pi,
		ForwardSpeed:  0,
	}
	h := yecs.Hierarchy{
		Parent: yecs.NullEntity,
	}

	w.AddComponent(ent, yecs.MeshEntryComponent, []yecs.MeshEntry{me})
	w.AddComponent(ent, yecs.TransformComponent, transform)
	w.AddComponent(ent, yecs.RenderStateComponent, renderState)
	w.AddComponent(ent, yecs.MoveComponent, move)
	w.AddComponent(ent, yecs.MaterialSurfaceComponent, *surface)
	w.AddComponent(ent, yecs.HierarchyComponent, h)
	return ent
}

func CreateCamera(w *yecs.World, follow yecs.EntityId) {
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
		//CamMode:        yecs.CAMERA_FOLLOW,
		Entity:         follow,
		SpringConstant: 5.0,
		Offset:         y3d.Vec3{X: 0.0, Y: 10, Z: 10},
		TargetDistance: y3d.Vec3{X: 10, Y: 10, Z: 10},
	}
	//camera.Recalulate()
	camera.SnapToIdeal(w)
	w.AddComponent(ent, yecs.CameraComponent, camera)
}

func randRange(min, max float32) float32 {
	return min + rand.Float32()*(max-min)
}

func CreateScene(w *yecs.World) {
	tex, err := ygl.CreateTex2D("assets/earth.jpg", gl.LINEAR_MIPMAP_LINEAR, gl.LINEAR, true)
	if err != nil {
		panic(err)
	}
	mesh := w.NewMesh()
	me := ygl.CreateSphereNew(56, 28, 1.0, mesh)
	mesh.Setup()
	surface := yecs.MaterialSurface{
		Diffuse: tex,
	}
	var e yecs.EntityId
	for i := range 1000 {
		x := randRange(-1000, 1000)
		y := randRange(-1000, 1000)
		z := randRange(10, 1000)
		transform := yecs.Transform{
			Position: y3d.Vec3{X: x, Y: y, Z: z},
			Scale:    y3d.Vec3{X: 64, Y: 64, Z: 64},
			Rotation: y3d.IdenQuat(),
		}
		(&transform).Recalulate()
		e = CreateObject(w, me, transform, i%3, &surface)
	}
	CreateCamera(w, e)

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
