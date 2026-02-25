package yam

import (
	"yam/y3d"
	"yam/ygame"
)

func NewTestAnim() {
	g, err := ygame.NewGame("test anim", width, height)
	if err != nil {
		panic(err)
	}
	err = g.ResourceManager.LoadSurfaceBundle("redhat", []string{"assets/redhatidle.png", "assets/redhatrun.png"})
	if err != nil {
		panic(err)
	}
	an := &ygame.AnimSprite{}
	an.Pos = y3d.Vec3{
		X: float64(g.Renderer.Width) / 2,
		Y: float64(g.Renderer.Height) / 2,
	}
	an.FrameHeight = 64
	an.FrameWidth = 64
	an.FPS = 10

	an.Texture, err = g.ResourceManager.GetSurface("redhat")
	if err != nil {
		panic(err)
	}
	an.SelectSheet(0)
	an.NumFrames = 10

	an2 := &ygame.AnimSprite{}
	an2.Pos = y3d.Vec3{
		X: float64(g.Renderer.Width)/2 + 128,
		Y: float64(g.Renderer.Height) / 2,
	}
	an2.FrameHeight = 64
	an2.FrameWidth = 64
	an2.FPS = 8

	an2.Texture, err = g.ResourceManager.GetSurface("redhat")
	if err != nil {
		panic(err)
	}
	an2.SelectSheet(1)
	an2.NumFrames = 8

	g.AddActor(an)
	g.AddActor(an2)
	g.Run()
}
