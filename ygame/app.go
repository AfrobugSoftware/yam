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
	dc := t.RenderManager.VertextManager.CreateDrawCommand(i, 1)
	staticbuf, err := t.RenderManager.VertextManager.CreateStaticBuffer(
		ycore.VP,
		v, i,
		[]ycore.DrawCommand{dc},
		skin,
		[]y3d.Mat4{y3d.Identity},
	)
	if err != nil {
		panic(err)
	}
	sp := ycore.NewSpatial(nil, y3d.AABB{},
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
	sp.Transform.Recalulate()
	sp.UpdateWorldTransform()
}

func (t *TestApplication) Update(currentTime float64) {}

func (t *TestApplication) Draw() {}

func (t *TestApplication) Shutdown() {}
