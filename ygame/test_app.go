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

	helmet, err := ycore.LoadGLTF("assets/gltf/DamagedHelmet.glb", t.engine.RenderManager)
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
	s, box := t.engine.RenderManager.VertextManager.GetScalingAndBox(v, i, 0.5, ycore.VP)
	sp := ycore.NewGeometry(
		t.engine.RenderManager,
		nil, box,
		ycore.NewTransform(),
		ycore.VP,
		v,
		i,
		skin,
		ycore.NO_STATICBUF, ycontroller.NewMovementController(
			y3d.Vec3{Z: -0.24}, y3d.IdenQuat(),
		))

	sp.Transform.SetScale(s)
	gt.Transform.SetScale(0.025)
	nt := ycore.NewNode(t.engine.RenderManager, nil, y3d.UnitAABB, ycore.NewTransform())
	nt.Add(sp)
	nt.Add(gt)
	nt.Add(helmet)

	t.Root = nt
	t.Root.UpdateWorldTransform()

	t.Obj = sp
	t.ObjPlayer = gt

	t.engine.RenderManager.FpCamera = ycore.NewCamera(y3d.ZEROV)
}

func (t *TestApplication) Update(deltaTime float64) {
	g := t.Obj.(*ycore.Geometry)
	if g.MoveController != nil {
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_PRESSED {
			g.MoveController.Thrust = -15
		}
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_PRESSED {
			g.MoveController.Thrust = 15
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_PRESSED {
			g.MoveController.RotSpeedRoll = -15
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_PRESSED {
			g.MoveController.RotSpeedRoll = 15
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_RELEASED {
			g.MoveController.Thrust = 0
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_RELEASED {
			g.MoveController.Thrust = 0
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_RELEASED {
			g.MoveController.RotSpeedRoll = 0
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_RELEASED {
			g.MoveController.RotSpeedRoll = 0
		}
	}

	cam := t.engine.RenderManager.FpCamera
	if cam != nil {
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_W) == ycore.BUTTON_PRESSED {
			cam.ForwardSpeed = 10
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_S) == ycore.BUTTON_PRESSED {
			cam.ForwardSpeed = -10
		}
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_W) == ycore.BUTTON_RELEASED {
			cam.ForwardSpeed = 0
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_S) == ycore.BUTTON_RELEASED {
			cam.ForwardSpeed = 0
		}
	}

	t.engine.RenderManager.UpdateFPCamera(float32(deltaTime))
	t.Root.UpdateControllers(float32(deltaTime))
	t.Root.UpdateWorldTransform()
}

func (t *TestApplication) Draw() {
	t.Root.Draw()
}

func (t *TestApplication) Shutdown() {}
