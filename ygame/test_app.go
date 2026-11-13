package ygame

import (
	"math"
	"yam/y3d"
	"yam/ycontroller"
	"yam/ycore"
	"yam/ygl"
	"yam/yui"

	"github.com/veandco/go-sdl2/sdl"
)

type TestApplication struct {
	engine       *Engine
	camFront     y3d.Vec3
	camPostiion  y3d.Vec3
	camUp        y3d.Vec3
	yaw, pitch   float32
	pickedEntity ycore.EntityId
	boxEnt       ycore.EntityId
	font         *yui.Font
}

func (t *TestApplication) Startup(e *Engine) {
	t.engine = e
	t.camFront = y3d.UNIT_Z
	v, i := ygl.CreateCube()
	if v == nil || i == nil {
		panic("failed to create cube")
	}

	err := ycore.LoadGLTF("assets/gltf/BrainStem.glb",
		t.engine.RenderManager,
		t.engine.Level.World)
	if err != nil {
		panic(err)
	}
	err = ycore.LoadGLTF("assets/gltf/SheenChair.glb",
		t.engine.RenderManager,
		t.engine.Level.World)
	if err != nil {
		panic(err)
	}
	t.camPostiion.Y = 1
	t.yaw = float32(y3d.ToRadians(90))
	t.camFront = y3d.Normalize(y3d.Vec3{
		X: float32(math.Cos(float64(t.yaw)) * math.Cos(float64(t.pitch))),
		Y: float32(math.Sin(float64(t.pitch))),
		Z: float32(math.Sin(float64(t.yaw)) * math.Cos(float64(t.pitch))),
	})
	t.font = yui.NewFont(12, t.engine.RenderManager.PixelPerInch,
		"assets/fonts/Mina-Regular.ttf")
	//t.TestUI() // test ui
	t.TestAbsolutePositioning()
	t.LevelTest() //level test
}
func (t *TestApplication) TestAbsolutePositioning() {
	root := yui.NewUIElement(
		y3d.Vec2{
			X: float32(t.engine.RenderManager.Width) - 210,
			Y: 10,
		},
		y3d.Vec2{
			X: 200,
			Y: 50,
		},
	).Set(yui.UI_ALIGN_CENTER).SetRadius(5).SetColor(
		[4]uint8{0, 0, 0, 255},
	).Add(
		&yui.UIElement{
			Flag:  yui.UI_ALIGN_CENTER,
			Text:  "Kiwi Island, this is a long text",
			Font:  t.font,
			Color: [4]uint8{255, 255, 255, 255},
		},
	)
	t.engine.RenderManager.RootUI = root
}

func (t *TestApplication) TestUI() {
	font := yui.NewFont(12, t.engine.RenderManager.PixelPerInch,
		"assets/fonts/Mina-Regular.ttf")
	root := yui.NewUIElement(y3d.Vec2{},
		y3d.Vec2{X: float32(t.engine.RenderManager.Width),
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
								Text:  "Zino is awesome",
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
	ent3 := t.engine.Level.World.NewEntity()
	ent4 := t.engine.Level.World.NewEntity()
	//skin := t.engine.RenderManager.SkinManager.AddSkin(ycore.IdentityMaterial)
	w := t.engine.Level.World
	m := ycore.AddEntityMoveData(t.engine.Level.World, ent, y3d.Vec3{Z: -20, Y: 5}, y3d.IdenQuat())
	m.Thrust = 0
	ycore.AddEntityMoveData(t.engine.Level.World, ent2, y3d.Vec3{Z: -50, Y: 5.0}, y3d.IdenQuat())
	ycore.AddEntityMoveData(t.engine.Level.World, ent4, y3d.Vec3{X: 5, Z: 10, Y: 5.0}, y3d.IdenQuat())

	v, i := ygl.CreateSphere(64, 36, 5.0)
	err := ycore.AddEntityRenderableData(t.engine.Level.World, ent,
		ycore.VPNTC, 1.0, v, i, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	v2, i2 := ygl.CreateCube()
	err = ycore.AddEntityRenderableData(t.engine.Level.World, ent3,
		ycore.VPNTC, 1.0, v2, i2, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	err = ycore.AddEntityRenderableData(t.engine.Level.World, ent2,
		ycore.VPNTC, 1.0, v, i, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	v, i = ygl.CreateQuad()
	err = ycore.AddEntityRenderableData(t.engine.Level.World, ent4,
		ycore.VPNTC, 1.0, v, i, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}

	ie1 := w.NewEntity()
	ie2 := w.NewEntity()
	ie3 := w.NewEntity()
	ycore.AddEntityMoveData(w, ie1, y3d.Vec3{X: 5, Z: 0, Y: 2.0}, y3d.IdenQuat())
	ycore.AddEntityMoveData(w, ie2, y3d.Vec3{X: 10, Z: 0, Y: 2.0}, y3d.IdenQuat())
	ycore.AddEntityMoveData(w, ie3, y3d.Vec3{X: -5, Z: 0, Y: 2.0}, y3d.IdenQuat())

	w.AddComponent(ie1, ycore.ScaleComponent, float32(35.0))
	w.AddComponent(ie2, ycore.ScaleComponent, float32(50.0))
	w.AddComponent(ie3, ycore.ScaleComponent, float32(20.0))
	ycore.AddEntityAsInstance(w, ent3, ie1)
	ycore.AddEntityAsInstance(w, ent3, ie2)
	ycore.AddEntityAsInstance(w, ent3, ie3)
	t.boxEnt = ent3
}

func (t *TestApplication) Update(deltaTime float64) {
	if gEngine.InputManager.GetKeyState(sdl.SCANCODE_P) == ycore.BUTTON_PRESSED {
		ycore.PrintMemoryUsage()
	}
	t.ProcessInput(float32(deltaTime))
	t.UpdateCamera()
}

func (t *TestApplication) ProcessInput(deltaTime float32) {
	camSpeed := 10 * deltaTime
	im := t.engine.InputManager
	caged := t.engine.InputManager.AbsolutePosition

	if im.GetMouseButtonState(sdl.BUTTON_LEFT) == ycore.BUTTON_HELD {
		if math.Abs(float64(im.MousePosition.X)) > y3d.NearZero ||
			math.Abs(float64(im.MousePosition.Y)) > y3d.NearZero {
			t.yaw += float32(y3d.ToRadians(float64(im.MousePosition.X))) * 0.5
			t.pitch += float32(y3d.ToRadians(float64(im.MousePosition.Y))) * 0.5
			if t.pitch > float32(y3d.ToRadians(89.0)) {
				t.pitch = float32(y3d.ToRadians(89.0))
			} else if t.pitch < -float32(y3d.ToRadians(89.0)) {
				t.pitch = -float32(y3d.ToRadians(89.0))
			}
			t.camFront = y3d.Normalize(y3d.Vec3{
				X: float32(math.Cos(float64(t.yaw)) * math.Cos(float64(t.pitch))),
				Y: float32(math.Sin(float64(t.pitch))),
				Z: float32(math.Sin(float64(t.yaw)) * math.Cos(float64(t.pitch))),
			})
		}
	}
	if im.GetMouseButtonState(sdl.BUTTON_LEFT) == ycore.BUTTON_PRESSED {
		pixels := t.engine.RenderManager.GetPickInfo(int32(caged.X), int32(caged.Y))
		b := t.engine.Level.World.IsEntity(ycore.EntityId(pixels[0]))
		if b {
			t.engine.RenderManager.PickedEntity = ycore.EntityId(pixels[0])
			t.pickedEntity = ycore.EntityId(pixels[0])
		} else {
			t.engine.RenderManager.PickedEntity = ycore.NullEntity
			t.pickedEntity = ycore.NullEntity
		}
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
	if im.GetKeyState(sdl.SCANCODE_B) == ycore.BUTTON_PRESSED {
		t.engine.RenderManager.DrawBoundingBoxes = !t.engine.RenderManager.DrawBoundingBoxes
	}
	if im.GetKeyState(sdl.SCANCODE_T) == ycore.BUTTON_PRESSED {
		t.engine.RenderManager.DrawBoundingTree = !t.engine.RenderManager.DrawBoundingTree
	}
	if im.GetKeyState(sdl.SCANCODE_L) == ycore.BUTTON_PRESSED {
		t.engine.RenderManager.DrawDebugLines = !t.engine.RenderManager.DrawDebugLines
	}
	if im.GetKeyState(sdl.SCANCODE_O) == ycore.BUTTON_PRESSED {
		t.engine.RenderManager.EnablePicking = !t.engine.RenderManager.EnablePicking
	}

	t.UpdatePickedEntity()
	t.AddBox()
}

func (t *TestApplication) AddBox() {
	im := t.engine.InputManager
	w := t.engine.Level.World
	caged := t.engine.InputManager.AbsolutePosition
	if im.GetMouseButtonState(sdl.BUTTON_RIGHT) == ycore.BUTTON_PRESSED {
		o := t.engine.RenderManager.GetPickRay(caged)
		ie1 := w.NewEntity()
		w.AddComponent(ie1, ycore.ScaleComponent, float32(35.0))
		ycore.AddEntityMoveData(w, ie1, y3d.Add(o.O, y3d.Smul(o.D, 5.0)), y3d.IdenQuat())
		ycore.AddEntityAsInstance(w, t.boxEnt, ie1)
	}
}

func (t *TestApplication) UpdatePickedEntity() {
	if t.pickedEntity == ycore.NullEntity {
		return
	}
	im := t.engine.InputManager
	w := t.engine.Level.World
	m := w.GetComponent(t.pickedEntity, ycore.MoveControllerComponent).(*ycontroller.MovementController)
	speed := float32(10)
	m.Thrust = 0
	if im.GetKeyState(sdl.SCANCODE_W) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_W) == ycore.BUTTON_HELD {
		m.Thrust = speed
	}
	if im.GetKeyState(sdl.SCANCODE_S) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_S) == ycore.BUTTON_HELD {
		m.Thrust = -speed
	}
	m.RotSpeedPitch = 0
	if im.GetKeyState(sdl.SCANCODE_A) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_A) == ycore.BUTTON_HELD {
		m.RotSpeedPitch = speed
	}
	if im.GetKeyState(sdl.SCANCODE_D) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_D) == ycore.BUTTON_HELD {
		m.RotSpeedPitch = -speed
	}
	m.RotSpeedRoll = 0
	if im.GetKeyState(sdl.SCANCODE_Q) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_Q) == ycore.BUTTON_HELD {
		m.RotSpeedRoll = speed
	}
	if im.GetKeyState(sdl.SCANCODE_E) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_E) == ycore.BUTTON_HELD {
		m.RotSpeedRoll = -speed
	}
	m.RotSpeedYaw = 0
	if im.GetKeyState(sdl.SCANCODE_Z) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_Z) == ycore.BUTTON_HELD {
		m.RotSpeedYaw = speed
	}
	if im.GetKeyState(sdl.SCANCODE_X) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_X) == ycore.BUTTON_HELD {
		m.RotSpeedYaw = -speed
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

func (t *TestApplication) Shutdown() {}
