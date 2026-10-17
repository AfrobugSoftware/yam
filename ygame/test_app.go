package ygame

import (
	"yam/y3d"
	"yam/ycontroller"
	"yam/ycore"
	"yam/ygl"
	"yam/yui"

	"github.com/veandco/go-sdl2/sdl"
)

type TestApplication struct {
	engine *Engine
}

func (t *TestApplication) Startup(e *Engine) {
	t.engine = e
	v, i := ygl.CreateCube()
	if v == nil || i == nil {
		panic("failed to create cube")
	}

	err := ycore.LoadGLTF("assets/gltf/DamagedHelmet.glb",
		t.engine.RenderManager,
		50, //tag
		t.engine.Level.World)
	if err != nil {
		panic(err)
	}
	t.engine.RenderManager.FpCamera = ycore.NewCamera(y3d.ZEROV)

	//t.TestUI() // test ui
	t.LevelTest() //level test
}

func (t TestApplication) TestUI() {
	font := yui.NewFont(12, t.engine.RenderManager.PixelPerInch,
		"assets/fonts/Mina-Regular.ttf")
	root := yui.NewUIElement(y3d.Vec2{X: float32(t.engine.RenderManager.Width),
		Y: float32(t.engine.RenderManager.Height)})
	root.Color = [4]uint8{39, 245, 0, 0}
	root.ChildGap = 20
	root.Padding = y3d.Vec4{X: 10, Y: 10, Z: 10, W: 10}
	root.Name = "root"
	leftP := &yui.UIElement{
		Name:     "left",
		BaseSize: y3d.Vec2{X: 150, Y: 0},
		Color:    [4]uint8{0, 255, 0, 255},
		Flag:     yui.UI_GROW_Y,
	}
	rightP := &yui.UIElement{
		Name:     "right",
		Color:    [4]uint8{0, 0, 255, 255},
		ChildGap: 40,
		Flag:     yui.UI_GROW_Y | yui.UI_GROW_X | yui.UI_LAYOUT_LEFT_TO_RIGHT,
		Padding:  y3d.Vec4{X: 10, Y: 10, Z: 10, W: 10},
		Children: []*yui.UIElement{
			{
				Name:     "inner-right",
				BaseSize: y3d.Vec2{X: 40, Y: 40},
				Flag:     yui.UI_ALIGN_CENTER,
				ChildGap: 12,
				Color:    [4]uint8{255, 255, 0, 0},
			},
			{
				Name:     "inner-right-container",
				BaseSize: y3d.Vec2{X: 40, Y: 40},
				Flag:     yui.UI_LAYOUT_TOP_TO_BOTTOM | yui.UI_GROW,
				Color:    [4]uint8{255, 255, 0, 255},
				Padding:  y3d.Vec4{X: 10, Y: 10, Z: 10, W: 10},
				ChildGap: 12,
				Children: []*yui.UIElement{
					{
						Name:     "inner-right-container-elem",
						Flag:     yui.UI_GROW | yui.UI_LAYOUT_TOP_TO_BOTTOM | yui.UI_ALIGN_CENTER,
						BaseSize: y3d.Vec2{X: 40, Y: 40},
						Color:    [4]uint8{245, 73, 39, 255},
						Radius:   25.0,
						Children: []*yui.UIElement{
							{
								Text:  "Hello world",
								Font:  font,
								Color: [4]uint8{255, 255, 0, 255},
							},
							{
								Text:  "This is a slighly longer text",
								Font:  font,
								Color: [4]uint8{255, 255, 0, 255},
							},
							{
								Text:  "Yram!!",
								Font:  font,
								Color: [4]uint8{255, 255, 0, 255},
							},
						},
					},
					{
						Name:     "inner-right-container-elem-2",
						Flag:     yui.UI_GROW,
						BaseSize: y3d.Vec2{X: 40, Y: 40},
						Color:    [4]uint8{242, 39, 245, 255},
					},
				},
			},
		},
	}

	root.Add(leftP, rightP)
	t.engine.RenderManager.RootUI = root
}

func (t *TestApplication) LevelTest() {
	ent := t.engine.Level.World.NewEntity()
	skin := t.engine.RenderManager.SkinManager.AddSkin(ygl.IdentityMaterial)
	v, i := ygl.CreateSphere(64, 36, 5.0)

	err := ycore.AddEntityDrawData(t.engine.Level.World, ent, 0x01,
		ycore.VPNTC, skin, y3d.UnitAABB, v, i, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	ycore.AddEntityMoveData(t.engine.Level.World, ent, y3d.Vec3{Z: -20}, y3d.IdenQuat())
	ycore.AddEntityInputData(t.engine.Level.World,
		ent,
		func(im *ycore.InputManager, c ycontroller.Controller) {
			if c == nil {
				return
			}
			m, ok := c.(*ycontroller.MovementController)
			if !ok {
				return
			}

			speed := 25.0
			if im.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_PRESSED {
				m.Thrust = -float32(speed)
			}
			if im.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_PRESSED {
				m.Thrust = float32(speed)
			}

			if im.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_PRESSED {
				m.RotSpeedRoll = float32(speed)
			}

			if im.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_PRESSED {
				m.RotSpeedRoll = -float32(speed)
			}
			if im.GetKeyState(sdl.SCANCODE_Z) == ycore.BUTTON_PRESSED {
				m.RotSpeedYaw = float32(speed)
			}

			if im.GetKeyState(sdl.SCANCODE_X) == ycore.BUTTON_PRESSED {
				m.RotSpeedYaw = -float32(speed)
			}

			if im.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_RELEASED {
				m.Thrust = 0
			}

			if im.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_RELEASED {
				m.Thrust = 0
			}

			if im.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_RELEASED {
				m.RotSpeedRoll = 0
			}

			if im.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_RELEASED {
				m.RotSpeedRoll = 0
			}
			if im.GetKeyState(sdl.SCANCODE_Z) == ycore.BUTTON_RELEASED {
				m.RotSpeedYaw = 0
			}

			if im.GetKeyState(sdl.SCANCODE_X) == ycore.BUTTON_RELEASED {
				m.RotSpeedYaw = 0
			}
		})
}

func (t *TestApplication) Update(deltaTime float64) {
	if gEngine.InputManager.GetKeyState(sdl.SCANCODE_P) == ycore.BUTTON_PRESSED {
		ycore.PrintMemoryUsage()
	}
	speed := 25.0

	cam := t.engine.RenderManager.FpCamera
	if cam != nil {
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_W) == ycore.BUTTON_PRESSED {
			cam.ForwardSpeed = -float32(speed)
		}
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_W) == ycore.BUTTON_RELEASED {
			cam.ForwardSpeed = 0
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_S) == ycore.BUTTON_PRESSED {
			cam.ForwardSpeed = float32(speed)
		}
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_S) == ycore.BUTTON_RELEASED {
			cam.ForwardSpeed = 0
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_A) == ycore.BUTTON_PRESSED {
			cam.RotSpeedYaw = -float32(speed)
		}
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_A) == ycore.BUTTON_RELEASED {
			cam.RotSpeedYaw = 0
		}

		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_D) == ycore.BUTTON_PRESSED {
			cam.RotSpeedYaw = float32(speed)
		}
		if gEngine.InputManager.GetKeyState(sdl.SCANCODE_D) == ycore.BUTTON_RELEASED {
			cam.RotSpeedYaw = 0
		}
	}

	t.engine.RenderManager.UpdateFPCamera(float32(deltaTime))
}

func (t *TestApplication) Draw() {

}

func (t *TestApplication) Shutdown() {}
