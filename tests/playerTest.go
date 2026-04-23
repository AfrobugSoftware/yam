package yam

import (
	"fmt"
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
	ygl.AssignUniformVec3(program, "light.diffuse", light.Diffuse.AsVec3())
	ygl.AssignUniformVec3(program, "light.ambient", light.Ambient.AsVec3())
	ygl.AssignUniformVec3(program, "light.specular", light.Specular.AsVec3())

	ygl.AssignUniformVec3(program, "material.diffuse", material.Diffuse.AsVec3())
	ygl.AssignUniformVec3(program, "material.ambient", material.Ambient.AsVec3())
	ygl.AssignUniformVec3(program, "material.specular", material.Specular.AsVec3())
	ygl.AssignUniformFloat32(program, "material.shininess", material.Shininess)
	return nil
}

func CreatePlayer(w *yecs.World) {
	e := w.NewEntity()
	renderState := yecs.RenderState{}
	renderState.States = append(renderState.States, yecs.DepthState{
		Enable:    true,
		DepthFunc: gl.LESS,
	}, yecs.BlendState{
		Enable:    true,
		SrcFactor: gl.SRC_ALPHA,
		DstFactor: gl.ONE_MINUS_SRC_ALPHA,
	})
	sprite := yecs.Spatial{
		Buffer:         "sphere",
		Program:        "simpleLight",
		CurTexture:     -1,
		AssignUniforms: AddUniforms,
	}
	transform := yecs.Transform{
		Position: y3d.Vec3{X: 0, Y: 0, Z: -100},
		Scale:    y3d.Vec3{X: 64, Y: 64, Z: 1},
		Rotation: y3d.IdenQuat(),
		Local:    y3d.Identity,
	}
	(&transform).Recalulate()
	move := yecs.Move{}
	in := ygame.GetGame().Input.CreateInput()
	in.Update = func(w *yecs.World, a yecs.EntityId) {
		input := w.GetComponent(e, yecs.InputComponent).(yecs.Input)
		move := yecs.Move{}

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

	material := yecs.Material{
		Shininess: 32,
		Diffuse: y3d.Color{
			1.0, 0.5, 0.31, 1.0,
		},
		Ambient: y3d.Color{
			1.0, 0.5, 0.31, 1.0,
		},
		Specular: y3d.Color{
			0.5, 0.5, 0.5, 1.0,
		},
	}

	w.AddComponent(e, yecs.TagComponent, tag)
	w.AddComponent(e, yecs.InputComponent, in)
	w.AddComponent(e, yecs.MoveComponent, move)
	w.AddComponent(e, yecs.SpatialComponent, sprite)
	w.AddComponent(e, yecs.MaterialComponent, material)
	w.AddComponent(e, yecs.TransformComponent, transform)
	w.AddComponent(e, yecs.RenderStateComponent, renderState)
}

func CreateLight(w *yecs.World) {
	e := w.NewEntity()
	light := yecs.Light{
		Pos:      y3d.Vec3{X: 100, Y: 10, Z: -20},
		Diffuse:  y3d.Color{0.5, 0.5, 0.5, 1.0},
		Ambient:  y3d.Color{0.2, 0.2, 0.2, 1.0},
		Specular: y3d.Color{1.0, 1.0, 1.0, 1.0},
	}
	w.AddComponent(e, yecs.LightComponent, light)
}

func TesPlayer() {
	g, err := ygame.NewGame("Test scene", width, height)
	if err != nil {
		panic(err)
	}
	CreateSystems(g.World)
	CreateResources(g)
	CreateLight(g.World)
	CreatePlayer(g.World)
	CreateCamera(g.World)
	g.Run()
}
