package ygame

import (
	"yam/y3d"
	"yam/ycore"
	"yam/ygl"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type Application interface {
	Startup()
	Update(currentTime float64)
	Draw()
	Shutdown()
}

type TestApplication struct {
	RenderManager *ycore.RenderManager
	Obj           ycore.SpatialInterface
}

func (t *TestApplication) Startup() {
	v, i := ygl.CreateCube()
	if v == nil || i == nil {
		panic("failed to create cube")
	}
	skin := t.RenderManager.SkinManager.AddSkin(ygl.IdentityMaterial)
	err := t.RenderManager.SkinManager.AddTexture(skin, "assets/earth.jpg",
		gl.LINEAR,
		gl.LINEAR,
		gl.CLAMP_TO_EDGE,
		gl.CLAMP_TO_EDGE,
		false)
	if err != nil {
		panic(err)
	}
	err = t.RenderManager.SkinManager.AddTexture(skin, "assets/earth.jpg",
		gl.LINEAR,
		gl.LINEAR,
		gl.CLAMP_TO_EDGE,
		gl.CLAMP_TO_EDGE,
		false)
	if err != nil {
		panic(err)
	}
	dc := t.RenderManager.VertextManager.CreateDrawCommand(i, 1)
	s, box := t.RenderManager.VertextManager.GetScalingAndBox(v, i, 0.5, ycore.VP)
	staticbuf, err := t.RenderManager.VertextManager.CreateStaticBuffer(
		ycore.VP,
		v, i,
		[]ycore.DrawCommand{dc},
		skin,
		1,
	)
	if err != nil {
		panic(err)
	}
	sp := ycore.NewGeometry(
		t.RenderManager,
		nil, box,
		ycore.NewTransform(),
		ycore.VP,
		v,
		i,
		dc,
		skin,
		staticbuf)
	t.RenderManager.Root = sp
	sp.Transform.Position = y3d.Vec3{
		X: 0.0,
		Y: 0.0,
		Z: -1.0,
	}
	sp.Transform.SetScale(s)
	sp.Transform.Recalulate()
	sp.UpdateWorldTransform()

	t.Obj = sp
}

func (t *TestApplication) Update(deltaTime float64) {
	speed := float32(0.5)
	trans := t.Obj.GetTransform()
	trans.Position = y3d.Add(trans.Position, y3d.Smul(y3d.NegateVec3(y3d.UNIT_Z), speed*float32(deltaTime)))

	trans.Recalulate()
	s := t.Obj.(*ycore.Geometry)
	s.UpdateWorldTransform()
}

func (t *TestApplication) Draw() {}

func (t *TestApplication) Shutdown() {}
