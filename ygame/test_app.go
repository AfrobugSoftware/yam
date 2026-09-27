package ygame

import (
	"yam/y3d"
	"yam/ycontroller"
	"yam/ycore"
	"yam/ygl"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

type TestApplication struct {
	engine    *Engine
	Obj       ycore.SpatialInterface
	ObjPlayer ycore.SpatialInterface
	Root      ycore.SpatialInterface
}

func (t *TestApplication) Startup(e *Engine) {
	t.engine = e
	v, i := ygl.CreateCube()
	if v == nil || i == nil {
		panic("failed to create cube")
	}

	gt, err := ycore.LoadGLTF("assets/gltf/testgltf.gltf", t.engine.RenderManager)
	if err != nil {
		panic(err)
	}
	skin := t.engine.RenderManager.SkinManager.AddSkin(ygl.IdentityMaterial)
	err = t.engine.RenderManager.SkinManager.AddTexture(skin, "assets/img/earth.jpg",
		gl.LINEAR,
		gl.LINEAR,
		gl.CLAMP_TO_EDGE,
		gl.CLAMP_TO_EDGE,
		false)
	if err != nil {
		panic(err)
	}
	err = t.engine.RenderManager.SkinManager.AddTexture(skin, "assets/img/earth.jpg",
		gl.LINEAR,
		gl.LINEAR,
		gl.CLAMP_TO_EDGE,
		gl.CLAMP_TO_EDGE,
		false)
	if err != nil {
		panic(err)
	}
	dc := t.engine.RenderManager.VertextManager.CreateDrawCommand(i, 1)
	s, box := t.engine.RenderManager.VertextManager.GetScalingAndBox(v, i, 0.5, ycore.VP)
	//USE STATIC BUFER
	// staticbuf, err := t.RenderManager.VertextManager.CreateStaticBuffer(
	// 	ycore.VP,
	// 	v, i,
	// 	[]ycore.DrawCommand{dc},
	// 	skin,
	// 	1,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	sp := ycore.NewGeometry(
		t.engine.RenderManager,
		nil, box,
		ycore.NewTransform(),
		ycore.VP,
		v,
		i,
		dc,
		skin,
		ycore.NO_STATICBUF, ycontroller.NewMovementController())

	sp.Transform.SetScale(s)
	gt.Transform.SetScale(0.025)
	nt := ycore.NewNode(t.engine.RenderManager, nil, y3d.UnitAABB, ycore.NewTransform())
	nt.Add(sp)
	nt.Add(gt)
	t.Root = nt
	t.Root.UpdateWorldTransform()

	t.Obj = sp
	t.ObjPlayer = gt
}

func (t *TestApplication) Update(deltaTime float64) {
	g := t.Obj.(*ycore.Geometry)
	if g.MoveController != nil {
		g.MoveController.Thrust = 0
		g.MoveController.RotSpeedRoll = 0
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_RELEASED || gEngine.InputManager.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_HELD {
			g.MoveController.Thrust = -15
		}
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_RELEASED || gEngine.InputManager.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_HELD {
			g.MoveController.Thrust = 15
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_RELEASED || gEngine.InputManager.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_HELD {
			g.MoveController.RotSpeedRoll = -15
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_RELEASED || gEngine.InputManager.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_HELD {
			g.MoveController.RotSpeedRoll = 15
		}
	}
	t.Root.UpdateControllers(float32(deltaTime))
	t.Root.UpdateWorldTransform()
}

func (t *TestApplication) Draw() {
	t.Root.Draw()
}

func (t *TestApplication) Shutdown() {}
