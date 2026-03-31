package yam

import (
	"yam/y3d"
	"yam/yecs"
	"yam/ygame"

	gl "github.com/chsc/gogl/gl33"
	"github.com/veandco/go-sdl2/sdl"
)

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
	sprite := yecs.Sprite{
		Buffer:     "sprite",
		Textures:   "sprite",
		Program:    "sprite",
		CurTexture: 0,
	}
	transform := yecs.Transform{
		Position:    y3d.Vec3{X: 0, Y: 0, Z: -100},
		Scale:       y3d.Vec3{X: 64, Y: 64, Z: 1},
		Orientation: y3d.IdenQuat(),
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
		w.SetComponent(e, yecs.MoveComponent, move)
	}

	w.AddComponent(e, yecs.InputComponent, in)
	w.AddComponent(e, yecs.SpriteComponent, sprite)
	w.AddComponent(e, yecs.TransformComponent, transform)
	w.AddComponent(e, yecs.RenderStateComponent, renderState)
	w.AddComponent(e, yecs.MoveComponent, move)
}

func TesPlayer() {
	g, err := ygame.NewGame("Test scene", width, height)
	if err != nil {
		panic(err)
	}
	CreateSystems(g.World)
	CreateResources(g)
	CreatePlayer(g.World)
	CreateCamera(g.World)
	g.Run()
}
