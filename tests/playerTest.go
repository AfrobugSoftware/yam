package yam

import (
	"fmt"
	"math"
	"yam/y3d"
	"yam/yecs"
	"yam/ygame"
	"yam/ygl"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

func AddUniforms(e yecs.EntityId, w *yecs.World, cam *yecs.Camera, program uint32) error {
	lightEntity := w.Query([]yecs.ComponentId{yecs.LightComponent})
	if len(lightEntity) == 0 {
		return fmt.Errorf("no light in scene")
	}
	light := w.GetComponent(lightEntity[0], yecs.LightComponent).(yecs.Light)
	material := w.GetComponent(e, yecs.MaterialComponent).(yecs.Material)

	ygl.AssignUniformVec3(program, "cameraPos", cam.Pos)

	ygl.AssignUniformVec3(program, "light.position", light.Pos)
	ygl.AssignUniformVec3(program, "light.diffuse", light.Diffuse)
	ygl.AssignUniformVec3(program, "light.ambient", light.Ambient)
	ygl.AssignUniformVec3(program, "light.specular", light.Specular)

	ygl.AssignUniformVec3(program, "material.diffuse", material.Diffuse)
	ygl.AssignUniformVec3(program, "material.ambient", material.Ambient)
	ygl.AssignUniformVec3(program, "material.specular", material.Specular)
	ygl.AssignUniformFloat32(program, "material.shininess", material.Shininess)
	return nil
}

func CreatePlayer(w *yecs.World, vbo ygl.VertBuffer) {
	e := w.NewEntity()
	renderState := yecs.RenderState{}
	renderState.States = append(renderState.States, yecs.DepthState{
		Enable:    true,
		DepthFunc: gl.LESS,
	}, yecs.BlendState{
		Enable:    false,
		SrcFactor: gl.SRC_ALPHA,
		DstFactor: gl.ONE_MINUS_SRC_ALPHA,
	})
	sprite := yecs.Spatial{
		Buf:            vbo.Buf,
		Indx:           vbo.Indx,
		VertArray:      vbo.VertArray,
		IndxCount:      vbo.IndxCount,
		Program:        0,
		CurTexture:     -1,
		AssignUniforms: AddUniforms,
	}
	transform := yecs.Transform{
		Position: y3d.Vec3{X: 0, Y: 0, Z: 100},
		Scale:    y3d.Vec3{X: 64, Y: 64, Z: 64},
		Rotation: y3d.IdenQuat(),
		Local:    y3d.Identity,
	}
	(&transform).Recalulate()
	move := yecs.Move{
		AnglularSpeed: 15 * math.Pi,
	}
	in := ygame.GetGame().Input.CreateInput()
	in.Update = func(w *yecs.World, a yecs.EntityId) {
		input := w.GetComponent(e, yecs.InputComponent).(yecs.Input)
		move := yecs.Move{
			AnglularSpeed: 15 * math.Pi,
		}

		if input.GetKeyState(sdl.SCANCODE_W) == yecs.BUTTON_PRESSED || input.GetKeyState(sdl.SCANCODE_W) == yecs.BUTTON_HELD {
			move.ForwardSpeed = 2000
		}
		if input.GetKeyState(sdl.SCANCODE_S) == yecs.BUTTON_PRESSED || input.GetKeyState(sdl.SCANCODE_S) == yecs.BUTTON_HELD {
			move.ForwardSpeed = -2000
		}
		if input.GetKeyState(sdl.SCANCODE_D) == yecs.BUTTON_PRESSED || input.GetKeyState(sdl.SCANCODE_D) == yecs.BUTTON_HELD {
			move.StrafeSpeed = 2000
		}
		if input.GetKeyState(sdl.SCANCODE_A) == yecs.BUTTON_PRESSED || input.GetKeyState(sdl.SCANCODE_A) == yecs.BUTTON_HELD {
			move.StrafeSpeed = -2000
		}
		if input.GetKeyState(sdl.SCANCODE_Q) == yecs.BUTTON_PRESSED || input.GetKeyState(sdl.SCANCODE_Q) == yecs.BUTTON_HELD {
			move.VerticalSpeed = 2000
		}
		if input.GetKeyState(sdl.SCANCODE_E) == yecs.BUTTON_PRESSED || input.GetKeyState(sdl.SCANCODE_E) == yecs.BUTTON_HELD {
			move.VerticalSpeed = -2000
		}
		w.SetComponent(e, yecs.MoveComponent, move)
	}

	tag := yecs.Tag{
		Name: "MainPlayer",
	}
	surface := yecs.MaterialSurface{}
	tex, err := ygl.CreateTex2D("assets/earth.jpg", gl.LINEAR_MIPMAP_LINEAR, gl.LINEAR, true)
	if err != nil {
		panic(err)
	}
	surface.Diffuse = tex
	h := yecs.Hierarchy{
		Parent: yecs.NullEntity,
	}
	w.AddComponent(e, yecs.TagComponent, tag)
	w.AddComponent(e, yecs.InputComponent, in)
	w.AddComponent(e, yecs.MoveComponent, move)
	w.AddComponent(e, yecs.SpatialComponent, sprite)
	w.AddComponent(e, yecs.TransformComponent, transform)
	w.AddComponent(e, yecs.RenderStateComponent, renderState)
	w.AddComponent(e, yecs.MaterialSurfaceComponent, surface)
	w.AddComponent(e, yecs.HierarchyComponent, h)
}

func CreateLight(w *yecs.World, count int) {
	for range count {
		e := w.NewEntity()
		x := randRange(-1000, 1000)
		y := randRange(-1000, 1000)
		z := randRange(-10, -1000)
		light := yecs.Light{
			Pos:      y3d.Vec3{X: x, Y: y, Z: z},
			Diffuse:  y3d.Vec3{X: 0.5, Y: 0.5, Z: 0.5},
			Ambient:  y3d.Vec3{X: 0.2, Y: 0.2, Z: 0.2},
			Specular: y3d.Vec3{X: 1.0, Y: 1.0, Z: 1.0},
		}
		w.AddComponent(e, yecs.LightComponent, light)
	}
}

func TesPlayer() {
	g, err := ygame.NewGame("Test scene", width, height)

	if err != nil {
		panic(err)
	}
	CreateSystems(g.World)
	CreateLight(g.World, 3)
	buffer, indices, format := ygl.CreateSphere(56, 28, 1.0)
	vbo := ygl.CreateVextexBuffer(buffer, indices, format)
	CreatePlayer(g.World, vbo)
	CreateCamera(g.World)

	g.Run()
}
