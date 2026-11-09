package yphy

import (
	"bytes"
	"unsafe"
	"yam/y3d"
)

type Sphere struct {
	Local y3d.Sphere
	World y3d.Sphere
}

func NewSpere() Sphere {
	return Sphere{}
}

func (s Sphere) Pick(r y3d.Ray) (bool, y3d.Vec3) {
	count, ts := r.IntersectsSphere(s.World)
	if count == 0 {
		return false, y3d.ZEROV3
	}
	if count == 1 {
		return true, y3d.Add(r.O, y3d.Smul(r.D, ts[0]))
	} else {
		t := min(ts[0], ts[1])
		return true, y3d.Add(r.O, y3d.Smul(r.D, t))
	}
}
func (s Sphere) GetColliderType() ColliderType {
	return COLLIDER_SPHERE
}

func (s Sphere) Intersects(c Collider) bool {
	return false
}

func BoundSphere(stride, relativeOffset int, dataV, dataI *bytes.Buffer) Sphere {
	bd := dataV.Bytes()
	getPos := func(i int) y3d.Vec3 {
		off := (int(stride) * i) + relativeOffset
		end := off + 3*int(unsafe.Sizeof(float32(0)))
		v := bd[off:end]
		vp := unsafe.Slice((*float32)(unsafe.Pointer(unsafe.SliceData(v))), 3)
		return y3d.Vec3{
			X: vp[0],
			Y: vp[1],
			Z: vp[2],
		}
	}
	count := dataI.Len() / 4
	i := dataI.Bytes()
	indices := unsafe.Slice((*uint32)(unsafe.Pointer(unsafe.SliceData(i))), count)
	x := getPos(int(indices[0]))
	y := getPos(int(indices[0]))
	maxDist := float32(-1)
	for _, i := range indices {
		v := getPos(int(i))
		d := y3d.Distance(x, v)
		if d > maxDist {
			maxDist = d
			y = v
		}
	}
	z := y
	maxDist = -1
	for _, i := range indices {
		v := getPos(int(i))
		d := y3d.Distance(y, v)
		if d > maxDist {
			maxDist = d
			z = v
		}
	}
	center := y3d.Vec3{
		X: (y.X + z.X) * 0.5,
		Y: (y.Y + z.Y) * 0.5,
		Z: (y.Z + z.Z) * 0.5,
	}
	radius := z.Length() * 0.5
	for _, i := range indices {
		v := getPos(int(i))
		d := y3d.Distance(center, v)
		if d > radius {
			oldRadius := radius
			radius = (oldRadius + d) * 0.5
			k := (radius - oldRadius) / d
			center = y3d.Vec3{
				X: center.X + k*(v.X-center.X),
				Y: center.Y + k*(v.Y-center.Y),
				Z: center.Z + k*(v.Z-center.Z),
			}
		}
	}
	return Sphere{
		Local: y3d.Sphere{
			C: center,
			R: radius,
		},
	}
}

func (s Sphere) RenderSphereMatrix() y3d.Mat4 {
	c := s.World.C
	r := s.World.R
	return y3d.Mat4{
		r, 0, 0, 0,
		0, r, 0, 0,
		0, 0, r, 0,
		c.X, c.Y, c.Z, 1.0,
	}
}

func (s Sphere) Transform(m y3d.Mat4) Collider {
	translate := y3d.Vec3{
		X: m[12],
		Y: m[13],
		Z: m[14],
	}
	c := y3d.Add(s.Local.C, translate)
	s.World.C = c
	s.World.R = s.Local.R
	return s
}
