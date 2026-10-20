package ygame

import (
	"math"
	"yam/y3d"
	"yam/ycore"
	"yam/ygl"
	"yam/yui"

	"github.com/veandco/go-sdl2/sdl"
)

type TestApplication struct {
	engine      *Engine
	camFront    y3d.Vec3
	camPostiion y3d.Vec3
	camUp       y3d.Vec3
	yaw, pitch  float32
}

func (t *TestApplication) Startup(e *Engine) {
	t.engine = e
	t.camFront = y3d.UNIT_Z
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
	err = ycore.LoadGLTF("assets/gltf/SheenChair.glb",
		t.engine.RenderManager,
		51, //tag
		t.engine.Level.World)
	if err != nil {
		panic(err)
	}
	t.engine.InputManager.IsRelative = true
	t.camPostiion.Y = 1
	t.yaw = y3d.TwoPI
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
	ent2 := t.engine.Level.World.NewEntity()
	skin := t.engine.RenderManager.SkinManager.AddSkin(ycore.IdentityMaterial)
	v, i := ygl.CreateSphere(64, 36, 5.0)

	err := ycore.AddEntityDrawData(t.engine.Level.World, ent, 0x01,
		ycore.VPNTC, skin, y3d.UnitAABB, v, i, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	err = ycore.AddEntityDrawData(t.engine.Level.World, ent2, 0x02,
		ycore.VPNTC, skin, y3d.UnitAABB, v, i, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	ycore.AddEntityMoveData(t.engine.Level.World, ent, y3d.Vec3{Z: -20}, y3d.IdenQuat())
	ycore.AddEntityMoveData(t.engine.Level.World, ent2, y3d.Vec3{Z: -50, Y: 5.0}, y3d.IdenQuat())

}

func (t *TestApplication) Update(deltaTime float64) {
	if gEngine.InputManager.GetKeyState(sdl.SCANCODE_P) == ycore.BUTTON_PRESSED {
		ycore.PrintMemoryUsage()
	}
	t.ProcessInput(float32(deltaTime))
	t.UpdateCamera()
}

func (t *TestApplication) ProcessInput(deltaTime float32) {
	camSpeed := 5 * deltaTime
	im := t.engine.InputManager

	if im.GetMouseButtonState(sdl.BUTTON_LEFT) == sdl.PRESSED {
		t.yaw += im.MousePosition.X * 0.01
		t.pitch += im.MousePosition.Y * 0.01
		if t.pitch > y3d.TwoPI {
			t.pitch -= y3d.TwoPI
		} else if t.pitch < -y3d.TwoPI {
			t.pitch += y3d.TwoPI
		}
		t.camFront = y3d.Normalize(y3d.Vec3{
			X: float32(math.Cos(float64(t.yaw)) * math.Cos(float64(t.pitch))),
			Y: float32(math.Sin(float64(t.pitch))),
			Z: float32(math.Sin(float64(t.yaw)) * math.Cos(float64(t.pitch))),
		})
	}
	if im.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_HELD {
		t.camPostiion = y3d.Add(t.camPostiion, y3d.Smul(t.camFront, float32(camSpeed)))
	}
	if im.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_HELD {
		t.camPostiion = y3d.Sub(t.camPostiion, y3d.Smul(t.camFront, float32(camSpeed)))
	}
	if im.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_HELD {
		right := y3d.Normalize(y3d.Cross(t.camFront, t.camUp))
		t.camPostiion = y3d.Add(t.camPostiion, y3d.Smul(right, float32(camSpeed)))
	}
	if im.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_HELD {
		right := y3d.Normalize(y3d.Cross(t.camFront, t.camUp))
		t.camPostiion = y3d.Sub(t.camPostiion, y3d.Smul(right, float32(camSpeed)))
	}
}

func (t *TestApplication) UpdateCamera() {
	dir := t.camFront
	vRight := y3d.Normalize(y3d.Cross(y3d.UNIT_Y, dir))
	vUp := y3d.Cross(dir, vRight)
	eye := t.camPostiion

	t.engine.RenderManager.View3D[t.engine.RenderManager.Stage] = y3d.Mat4{
		vRight.X, vUp.X, dir.X, 0,
		vRight.Y, vUp.Y, dir.Y, 0,
		vRight.Z, vUp.Z, dir.Z, 0,
		-y3d.Dot(vRight, eye), -y3d.Dot(vUp, eye), -y3d.Dot(dir, eye), 1,
	}
	t.camUp = vUp
}
func (t *TestApplication) Draw() {

}

func (t *TestApplication) Shutdown() {}
