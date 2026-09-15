package ygame

import (
	"yam/y3d"
	"yam/ygl"
	"yam/ymanager"
)

type Application interface {
	Startup()
	Update(currentTime float64)
	Draw()
	Shutdown()
}

type TestApplication struct {
	RenderManager *ymanager.RenderManager
}

func (t *TestApplication) Startup() {
	v, i := ygl.CreateCube()
	if v == nil || i == nil {
		panic("failed to create cube")
	}
	dc := t.RenderManager.VertextManager.CreateDrawCommand(i, 1)
	s, err := t.RenderManager.VertextManager.CreateStaticBuffer(
		ymanager.VP,
		v, i,
		[]ymanager.DrawCommand{dc},
		-1,
		[]y3d.Mat4{y3d.Translation(y3d.Vec3{
			X: 0.0,
			Y: 0.0,
			Z: -10.0,
		})},
	)
	if err != nil {
		panic(err)
	}
	sp := ymanager.NewSpatial(nil, y3d.AABB{},
		ymanager.NewTransform(),
		ymanager.VP,
		nil, nil,
		dc,
		-1,
		s)
	t.RenderManager.Root = sp
}

func (t *TestApplication) Update(currentTime float64) {}

func (t *TestApplication) Draw() {}

func (t *TestApplication) Shutdown() {}
