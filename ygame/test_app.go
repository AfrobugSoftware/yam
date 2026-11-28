package ygame

import (
	"math"
	"yam/y3d"
	"yam/ycore"
	"yam/ygl"
	"yam/yphy"
	"yam/yui"
	"yam/ywd"

	"github.com/veandco/go-sdl2/sdl"
)

type TestApplication struct {
	engine        *Engine
	camFront      y3d.Vec3
	camPostiion   y3d.Vec3
	camUp         y3d.Vec3
	yaw, pitch    float32
	pickedEntity  ywd.EntityId
	boxEnt        ywd.EntityId
	sphereEnt     ywd.EntityId
	font          *yui.Font
	useCam        bool
	drag          *yphy.DragForce
	forceRegistry *yphy.ForceRegistry
}

func (t *TestApplication) Startup(e *Engine) {
	t.engine = e
	t.useCam = true
	t.camFront = y3d.UNIT_Z

	//loading these things take so much memory, like 200mb worth
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
	err = ycore.LoadGLTF("assets/gltf/DamagedHelmet.glb",
		t.engine.RenderManager,
		t.engine.Level.World)
	if err != nil {
		panic(err)
	}
	// err = ycore.LoadGLTF("assets/gltf/Fox.glb",
	// 	t.engine.RenderManager,
	// 	t.engine.Level.World)
	// if err != nil {
	// 	panic(err)
	// }
	t.forceRegistry = yphy.NewForceRegistry()
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
	root := yui.Column(y3d.Vec4{X: 2, Y: 2, Z: 2, W: 2},
		10,
		yui.Button("Click me", y3d.Vec2{
			X: 110,
			Y: 50,
		}, [4]uint8{0, 0, 0, 255}, t.font),
		yui.SizedBox(y3d.Vec2{X: 0, Y: 10}),
		yui.Button("Click me again", y3d.Vec2{
			X: 150,
			Y: 50,
		}, yui.BLUE, t.font),
	)
	t.engine.RenderManager.RootUI = root
	t.drag = yphy.NewDragForce(0.5, 0.5)
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
	w := t.engine.Level.World
	ycore.AddEntityMoveData(t.engine.Level.World, ent, y3d.Vec3{Z: -20, Y: 5}, y3d.IdenQuat())
	ycore.AddEntityMoveData(t.engine.Level.World, ent2, y3d.Vec3{Z: -50, Y: 5.0}, y3d.IdenQuat())
	ycore.AddEntityMoveData(t.engine.Level.World, ent4, y3d.Vec3{X: 5, Z: 10, Y: 5.0}, y3d.IdenQuat())

	v, i := ygl.CreateSphere(64, 36, 1.0)
	err := ycore.AddEntityRenderableData(t.engine.Level.World, ent,
		ycore.VPNTC, y3d.Vec3{X: 1.0, Y: 1.0, Z: 1.0}, v, i, nil, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	t.sphereEnt = ent
	v2, i2 := ygl.CreateCube()
	err = ycore.AddEntityRenderableData(t.engine.Level.World, ent3,
		ycore.VPNTC, y3d.Vec3{X: 1.0, Y: 1.0, Z: 1.0}, v2, i2, nil, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	err = ycore.AddEntityRenderableData(t.engine.Level.World, ent2,
		ycore.VPNTC, y3d.Vec3{X: 1.0, Y: 1.0, Z: 1.0}, v, i, nil, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}
	v, i = ygl.CreateQuad()
	err = ycore.AddEntityRenderableData(t.engine.Level.World, ent4,
		ycore.VPNTC, y3d.Vec3{X: 1.0, Y: 1.0, Z: 1.0}, v, i, nil, t.engine.RenderManager.VertextManager)
	if err != nil {
		panic(err)
	}

	ie1 := w.NewEntity()
	ie2 := w.NewEntity()
	ie3 := w.NewEntity()
	ycore.AddEntityMoveData(w, ie1, y3d.Vec3{X: 5, Z: 0, Y: 2.0}, y3d.IdenQuat())
	ycore.AddEntityMoveData(w, ie2, y3d.Vec3{X: 10, Z: 0, Y: 2.0}, y3d.IdenQuat())
	ycore.AddEntityMoveData(w, ie3, y3d.Vec3{X: -5, Z: 0, Y: 2.0}, y3d.IdenQuat())
	unitScale := y3d.Vec3{X: 1, Y: 1, Z: 1}
	w.AddComponent(ie1, ywd.ScaleComponent, unitScale)
	w.AddComponent(ie2, ywd.ScaleComponent, unitScale)
	w.AddComponent(ie3, ywd.ScaleComponent, unitScale)
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

	t.UpdatePickedEntity()
	t.Scale()
	t.AddBox()
	t.AddSphere()
	t.ShotBullet()
	t.forceRegistry.Update(t.engine.RenderManager.World,
		float32(deltaTime))
	t.UpdateCamera()
}

func (t *TestApplication) ProcessInput(deltaTime float32) {
	camSpeed := 10 * deltaTime
	im := t.engine.InputManager
	caged := t.engine.InputManager.AbsolutePosition

	if im.GetMouseButtonState(sdl.BUTTON_LEFT) == ycore.BUTTON_HELD {
		if t.useCam {
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
		} else {
			if t.pickedEntity != ywd.NullEntity {
				t.engine.RenderManager.DragObject(t.pickedEntity, caged)
			}
		}
	}
	if im.GetMouseButtonState(sdl.BUTTON_LEFT) == ycore.BUTTON_PRESSED {
		pixels := t.engine.RenderManager.GetPickInfo(int32(caged.X), int32(caged.Y))
		b := t.engine.Level.World.IsEntity(ywd.EntityId(pixels[0]))
		if b {
			t.engine.RenderManager.PickedEntity = ywd.EntityId(pixels[0])
			t.pickedEntity = ywd.EntityId(pixels[0])
		} else {
			t.engine.RenderManager.PickedEntity = ywd.NullEntity
			t.pickedEntity = ywd.NullEntity
		}
	}
	if im.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_HELD {
		t.camPostiion = t.camPostiion.AddScale(t.camFront, -camSpeed)
	}
	if im.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_HELD {
		t.camPostiion = t.camPostiion.AddScale(t.camFront, camSpeed)
	}
	if im.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_RIGHT) == ycore.BUTTON_HELD {
		right := y3d.Normalize(y3d.Cross(t.camFront, t.camUp))
		t.camPostiion = t.camPostiion.AddScale(right, camSpeed)
	}
	if im.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_LEFT) == ycore.BUTTON_HELD {
		right := y3d.Normalize(y3d.Cross(t.camFront, t.camUp))
		t.camPostiion = t.camPostiion.AddScale(right, -camSpeed)
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
	if im.GetKeyState(sdl.SCANCODE_C) == ycore.BUTTON_PRESSED {
		t.useCam = !t.useCam
	}

}

func (t *TestApplication) CreateGroundPlane() {
	w := t.engine.Level.World
	ie1 := w.NewEntity()
	w.AddComponent(ie1, ywd.PositionComponent, y3d.ZEROV3)
	w.AddComponent(ie1, ywd.OrientationComponent, y3d.IdenQuat())
	w.AddComponent(ie1, ywd.VelocityComponent, y3d.Vec3{})
	w.AddComponent(ie1, ywd.InverseMassComponent, float32(0.0))
	w.AddComponent(ie1, ywd.AccelerationComponent, y3d.Vec3{})
	w.AddComponent(ie1, ywd.ColliderComponent, yphy.Plane{
		Local: y3d.PlaneFromPointNormal(y3d.ZEROV3, y3d.UNIT_Y),
		World: y3d.PlaneFromPointNormal(y3d.ZEROV3, y3d.UNIT_Y),
	})
}

func (t *TestApplication) AddBox() {
	im := t.engine.InputManager
	w := t.engine.Level.World
	caged := t.engine.InputManager.AbsolutePosition
	if im.GetMouseButtonState(sdl.BUTTON_RIGHT) == ycore.BUTTON_PRESSED {
		o := t.engine.RenderManager.GetPickRay(caged)
		ie1 := w.NewEntity()
		w.AddComponent(ie1, ywd.TransformComponent, ycore.NewTransform())
		ycore.AddEntityAsInstance(w, t.boxEnt, ie1)
		w.AddComponent(ie1, ywd.PositionComponent, o.O)
		w.AddComponent(ie1, ywd.OrientationComponent, y3d.IdenQuat())
		w.AddComponent(ie1, ywd.ScaleComponent, y3d.Vec3{X: 0.25, Y: 0.25, Z: 0.25})
		w.AddComponent(ie1, ywd.VelocityComponent, y3d.Vec3{})
		w.AddComponent(ie1, ywd.InverseMassComponent, float32(1.0/2.0))
		w.AddComponent(ie1, ywd.MoveComponent, &ycore.Move{})
		w.AddComponent(ie1, ywd.AccelerationComponent, y3d.Vec3{
			X: 0.0,
			Y: -1.0,
			Z: 0.0,
		})
		w.AddComponent(ie1, ywd.ForceComponent, y3d.Vec3{})
	}
}

func (t *TestApplication) AddSphere() {
	im := t.engine.InputManager
	w := t.engine.Level.World
	caged := t.engine.InputManager.AbsolutePosition
	if im.GetKeyState(sdl.SCANCODE_I) == ycore.BUTTON_PRESSED {
		o := t.engine.RenderManager.GetPickRay(caged)
		ie1 := w.NewEntity()
		w.AddComponent(ie1, ywd.TransformComponent, ycore.NewTransform())
		ycore.AddEntityAsInstance(w, t.sphereEnt, ie1)
		w.AddComponent(ie1, ywd.PositionComponent, o.O)
		w.AddComponent(ie1, ywd.OrientationComponent, y3d.IdenQuat())
		w.AddComponent(ie1, ywd.ScaleComponent, y3d.Vec3{X: 0.25, Y: 0.25, Z: 0.25})
		w.AddComponent(ie1, ywd.VelocityComponent, y3d.Smul(o.D, 15))
		w.AddComponent(ie1, ywd.InverseMassComponent, float32(1.0/2.0))
		w.AddComponent(ie1, ywd.MoveComponent, &ycore.Move{})
		w.AddComponent(ie1, ywd.AccelerationComponent, y3d.Vec3{
			X: 0.0,
			Y: -1.0,
			Z: 0.0,
		})
		w.AddComponent(ie1, ywd.ForceComponent, y3d.Vec3{})
	}
}

func (t *TestApplication) ShotBullet() {
	im := t.engine.InputManager
	w := t.engine.Level.World
	caged := t.engine.InputManager.AbsolutePosition
	if im.GetKeyState(sdl.SCANCODE_O) == ycore.BUTTON_PRESSED {
		o := t.engine.RenderManager.GetPickRay(caged)
		ie1 := w.NewEntity()
		w.AddComponent(ie1, ywd.TransformComponent, ycore.NewTransform())
		ycore.AddEntityAsInstance(w, t.boxEnt, ie1)
		w.AddComponent(ie1, ywd.PositionComponent, o.O)
		w.AddComponent(ie1, ywd.OrientationComponent, y3d.IdenQuat())
		w.AddComponent(ie1, ywd.ScaleComponent, y3d.Vec3{X: 0.25, Y: 0.25, Z: 0.25})
		w.AddComponent(ie1, ywd.VelocityComponent, y3d.Smul(o.D, 35))
		w.AddComponent(ie1, ywd.InverseMassComponent, float32(1.0/2.0))
		w.AddComponent(ie1, ywd.AccelerationComponent, y3d.Vec3{
			X: 0.0,
			Y: -1.0,
			Z: 0.0,
		})
		w.AddComponent(ie1, ywd.ForceComponent, y3d.Vec3{})
	}
}

func (t *TestApplication) Scale() {
	im := t.engine.InputManager
	w := t.engine.Level.World
	if t.pickedEntity != ywd.NullEntity {
		scale := w.GetComponent(t.pickedEntity, ywd.ScaleComponent).(y3d.Vec3)
		if im.GetKeyState(sdl.SCANCODE_KP_PLUS) == ycore.BUTTON_PRESSED {
			scale = y3d.Add(scale, y3d.Vec3{X: 0.25, Y: 0.25, Z: 0.25})
		}
		if im.GetKeyState(sdl.SCANCODE_KP_MINUS) == ycore.BUTTON_PRESSED {
			scale = y3d.Sub(scale, y3d.Vec3{X: 0.25, Y: 0.25, Z: 0.25})
			//need to clamp to zero if not itwould shear or flip
		}
		w.SetComponent(t.pickedEntity, ywd.ScaleComponent, scale)
	}
}

func (t *TestApplication) UpdatePickedEntity() {
	if t.pickedEntity == ywd.NullEntity {
		return
	}

	im := t.engine.InputManager
	w := t.engine.Level.World
	m := w.GetComponent(t.pickedEntity, ywd.MoveComponent).(*ycore.Move)
	speed := float32(10)
	m.ForwardSpeed = 0
	if im.GetKeyState(sdl.SCANCODE_W) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_W) == ycore.BUTTON_HELD {
		m.ForwardSpeed = speed
	}
	if im.GetKeyState(sdl.SCANCODE_S) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_S) == ycore.BUTTON_HELD {
		m.ForwardSpeed = -speed
	}

	m.RightSpeed = 0
	if im.GetKeyState(sdl.SCANCODE_A) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_A) == ycore.BUTTON_HELD {
		m.RightSpeed = -speed
	}
	if im.GetKeyState(sdl.SCANCODE_D) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_D) == ycore.BUTTON_HELD {
		m.RightSpeed = speed
	}
	m.RotSpeedPitch = 0
	if im.GetKeyState(sdl.SCANCODE_U) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_U) == ycore.BUTTON_HELD {
		m.RotSpeedPitch = speed
	}
	if im.GetKeyState(sdl.SCANCODE_Y) == ycore.BUTTON_PRESSED || im.GetKeyState(sdl.SCANCODE_Y) == ycore.BUTTON_HELD {
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
