package ygame

import (
	"yam/y3d"
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
		ycore.NO_STATICBUF)

	sp.Transform.Position = y3d.Vec3{
		X: 0.0,
		Y: 0.0,
		Z: -1.0,
	}
	sp.Transform.SetScale(s)
	sp.Transform.Recalulate()
	sp.UpdateWorldTransform()
	gt.Transform.Position = y3d.Vec3{
		X: 0.0,
		Y: 0.0,
		Z: -0.25,
	}
	gt.Transform.SetScale(0.025)
	gt.Transform.Recalulate()
	gt.UpdateWorldTransform()

	nt := ycore.NewNode(t.engine.RenderManager, nil, y3d.UnitAABB, ycore.NewTransform())
	nt.Add(sp)
	nt.Add(gt)
	t.Root = nt

	t.Obj = sp
	t.ObjPlayer = gt
}

func (t *TestApplication) Update(deltaTime float64) {
	speed := float32(0.5)
	trans := t.Obj.GetTransform()
	trans.Position = y3d.Add(trans.Position, y3d.Smul(y3d.NegateVec3(y3d.UNIT_Z), speed*float32(deltaTime)))

	trans.Recalulate()
	t.Obj.UpdateWorldTransform()

	speed = 0
	trans = t.ObjPlayer.GetTransform()
	if gEngine.InputManager.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_RELEASED || gEngine.InputManager.GetKeyState(sdl.SCANCODE_UP) == ycore.BUTTON_HELD {
		speed = 5
	}
	if gEngine.InputManager.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_RELEASED || gEngine.InputManager.GetKeyState(sdl.SCANCODE_DOWN) == ycore.BUTTON_HELD {
		speed = -5
	}
	trans.Position = y3d.Add(trans.Position, y3d.Smul(y3d.UNIT_Y, speed*float32(deltaTime)))
	trans.Recalulate()
	t.ObjPlayer.UpdateWorldTransform()
}

func (t *TestApplication) Draw() {
	t.Root.Draw()
}

func (t *TestApplication) Shutdown() {}
