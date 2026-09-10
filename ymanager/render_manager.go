package ymanager

import (
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	MODE_2D             = 0
	MODE_3D_ORTHO       = 1
	MODE_3D_PERSPECTIVE = 2
)

type RenderManager struct {
	ViewPort      [4]y3d.Rect
	View2D        y3d.Mat4
	View3D        y3d.Mat4
	Proj2D        y3d.Mat4
	ProjP         [4]y3d.Mat4
	ProjO         [4]y3d.Mat4
	World         y3d.Mat4
	ViewProj      y3d.Mat4
	WorldViewProj y3d.Mat4
	Near          float32
	Far           float32
	Width         int
	Height        int
	Fov           float32
	AspectRatio   float32
	Stage         int
	Mode          int
	SkinManager   *SkinManager
}

func (r *RenderManager) GetFrustum() [6]y3d.Plane {
	var planes [6]y3d.Plane
	//left
	planes[0] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] + r.ViewProj[0]),
			Y: -(r.ViewProj[13] + r.ViewProj[1]),
			Z: -(r.ViewProj[14] + r.ViewProj[2]),
		},
		D: -(r.ViewProj[15] + r.ViewProj[3]),
	}
	//right
	planes[1] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] - r.ViewProj[0]),
			Y: -(r.ViewProj[13] - r.ViewProj[1]),
			Z: -(r.ViewProj[14] - r.ViewProj[2]),
		},
		D: -(r.ViewProj[15] - r.ViewProj[3]),
	}
	//top
	planes[2] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] - r.ViewProj[4]),
			Y: -(r.ViewProj[13] - r.ViewProj[5]),
			Z: -(r.ViewProj[14] - r.ViewProj[6]),
		},
		D: -(r.ViewProj[15] - r.ViewProj[7]),
	}
	//bottom
	planes[3] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] + r.ViewProj[4]),
			Y: -(r.ViewProj[13] + r.ViewProj[5]),
			Z: -(r.ViewProj[14] + r.ViewProj[6]),
		},
		D: -(r.ViewProj[15] + r.ViewProj[7]),
	}
	//near
	planes[4] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[8]),
			Y: -(r.ViewProj[9]),
			Z: -(r.ViewProj[10]),
		},
		D: -(r.ViewProj[11]),
	}
	//bottom
	planes[5] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] - r.ViewProj[8]),
			Y: -(r.ViewProj[13] - r.ViewProj[9]),
			Z: -(r.ViewProj[14] - r.ViewProj[10]),
		},
		D: -(r.ViewProj[15] - r.ViewProj[11]),
	}
	for i := range planes {
		l := planes[i].N.Length()
		planes[i].D /= l
		planes[i].N = y3d.SDiv(planes[i].N, l)
	}
	return planes
}

func (r *RenderManager) SetClippingPlanes(near, far float32) {
	if near > far {
		near = far
		far = near + 1
	}
	if near <= 0.0 {
		near = 0.01
	}
	if far <= 1.0 {
		far = 1.00
	}
	r.Near = near
	r.Far = far
	r.Prepare2D()

	Q := 1.0 / (r.Far - r.Near)
	X := -Q * r.Near
	r.ProjO[0][10] = Q
	r.ProjO[1][10] = Q
	r.ProjO[2][10] = Q
	r.ProjO[3][10] = Q
	r.ProjO[0][14] = X
	r.ProjO[1][14] = X
	r.ProjO[2][14] = X
	r.ProjO[3][14] = X

	Q *= r.Far
	X = -Q * r.Near
	r.ProjP[0][10] = Q
	r.ProjP[1][10] = Q
	r.ProjP[2][10] = Q
	r.ProjP[3][10] = Q
	r.ProjP[0][14] = X
	r.ProjP[1][14] = X
	r.ProjP[2][14] = X
	r.ProjP[3][14] = X
}

func NewRenderManager() *RenderManager {
	rm := &RenderManager{
		View2D: y3d.Identity,
	}
	return rm
}

func (r *RenderManager) Prepare2D() {
	r.View2D = y3d.Identity
	r.Proj2D[0] = 2.0 / float32(r.Width)
	r.Proj2D[5] = 2.0 / float32(r.Height)
	r.Proj2D[10] = 1.0 / (r.Far - r.Near)
	r.Proj2D[11] = -r.Near * (1.0 / (r.Far - r.Near))
	r.Proj2D[15] = 1.0
	//flip the y axis for 2D rendering``
	r.View2D[5] = -1.0
	r.View2D[12] = -float32(r.Width) + float32(r.Width)*0.5
	r.View2D[13] = float32(r.Height) - float32(r.Height)*0.5
	r.View2D[14] = r.Near + 0.1
}

func (r *RenderManager) CalcViewProj() {
	var a, b *y3d.Mat4
	switch r.Mode {
	case MODE_2D:
		a = &r.Proj2D
		b = &r.View2D
	case MODE_3D_ORTHO:
		a = &r.ProjO[r.Stage]
		b = &r.View3D
	case MODE_3D_PERSPECTIVE:
		a = &r.ProjP[r.Stage]
		b = &r.View3D
	default:
		a = &y3d.Identity
		b = &y3d.Identity
	}
	r.ViewProj = a.Mul(*b)
}

func (r *RenderManager) CalcWorldViewProj() {
	r.WorldViewProj = r.ViewProj.Mul(r.World)
}

func (r *RenderManager) SetStage(mode int, stage int) {
	if stage < 0 || stage >= 4 {
		stage = 0
	}
	if r.Mode != mode || r.Stage != stage {
		r.Mode = mode
		r.Stage = stage
	}
	gl.Viewport(int32(r.ViewPort[stage].X),
		int32(r.ViewPort[stage].Y),
		int32(r.ViewPort[stage].Width),
		int32(r.ViewPort[stage].Height))
	r.CalcViewProj()
	r.CalcWorldViewProj()
}

func (r *RenderManager) InitStage(mode int, stage int, viewport y3d.Rect, fov float32) {
	if stage < 0 || stage >= 4 {
		stage = 0
	}
	r.ViewPort[stage] = viewport
	r.Fov = fov
	r.AspectRatio = float32(viewport.Width) / float32(viewport.Height)
	r.ProjP[stage] = y3d.Perspective(fov, r.AspectRatio, r.Near, r.Far)
	r.ProjO[stage] = y3d.Identity

	r.ProjO[stage][0] = 2 / float32(viewport.Width)
	r.ProjO[stage][5] = 2 / float32(viewport.Height)
	r.ProjO[stage][10] = 1.0 / (r.Far - r.Near)
	r.ProjO[stage][11] = -r.Near * (1.0 / (r.Far - r.Near))
	r.ProjO[stage][15] = 1.0
}

func (r *RenderManager) Transfrom3Dto2D(pos y3d.Vec3) y3d.Vec2 {
	var width, height float32
	if r.Mode == MODE_2D {
		width, height = float32(r.Width), float32(r.Height)
	} else {
		width, height = float32(r.ViewPort[r.Stage].Width), float32(r.ViewPort[r.Stage].Height)
	}
	clipX := width / 2
	clipY := height / 2

	fXp := r.ViewProj.MulVec4(y3d.Vec4{X: pos.X, Y: pos.Y, Z: pos.Z, W: 1.0})
	Winv := 1.0 / fXp.W
	fXp.X *= Winv
	fXp.Y *= Winv
	fXp.Z *= Winv
	fXp.W = 1.0
	return y3d.Vec2{
		X: (1.0 + fXp.X) * clipX,
		Y: (1.0 + fXp.Y) * clipY,
	}
}

func (r *RenderManager) Transfrom2Dto3D(pos y3d.Vec2) (y3d.Vec3, y3d.Vec3) {
	var width, height float32
	if r.Mode == MODE_2D {
		width, height = float32(r.Width), float32(r.Height)
	} else {
		width, height = float32(r.ViewPort[r.Stage].Width), float32(r.ViewPort[r.Stage].Height)
	}
	vcS := y3d.Vec3{
		X: (((pos.X * 2.0) / width) - 1.0) / r.ProjP[r.Stage][0],
		Y: (((pos.Y * 2.0) / height) - 1.0) / r.ProjP[r.Stage][5],
		Z: 1.0,
	}
	vcS.Y = -vcS.Y
	InvView := r.View3D
	(&InvView).Invert()
	var dir, origin y3d.Vec3
	dir = InvView.MulVec3(vcS)
	origin = y3d.Vec3{
		X: InvView[12],
		Y: InvView[13],
		Z: InvView[14],
	}
	dir = y3d.Normalize(dir)
	return origin, dir
}
