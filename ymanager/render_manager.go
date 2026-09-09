package ymanager

import "yam/y3d"

const (
	MODE_2D = 0
	MODE_3D
)

type RenderManager struct {
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

	Mode        int
	SkinManager *SkinManager
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

}
