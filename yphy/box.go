package yphy

import (
	"bytes"
	"unsafe"
	"yam/y3d"
)

type Box struct {
	Local y3d.AABB
	World y3d.AABB
}

func NewBox() Box {
	return Box{
		Local: y3d.UnitAABB,
		World: y3d.UnitAABB,
	}
}
func (b Box) Pick(ray y3d.Ray) (bool, y3d.Vec3) {
	return ray.IntersectsAABB(b.World)
}

func (b Box) Transform(m y3d.Mat4) Collider {
	translate := y3d.Vec3{
		X: m[12],
		Y: m[13],
		Z: m[14],
	}
	max := y3d.Add(b.Local.Max, translate)
	min := y3d.Add(b.Local.Min, translate)
	b.World = y3d.AABB{
		Min: min,
		Max: max,
	}
	return b
}

func (b Box) GetColliderType() ColliderType {
	return COLLIDER_AABB
}

func (b Box) Intersects(c Collider) bool {
	return false
}

func BoundBox(stride, relativeOffset int, dataV, dataI *bytes.Buffer) Box {
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
	box := y3d.AABB{
		Min: y3d.Vec3{
			X: 999999.99,
			Y: 999999.99,
			Z: 999999.99,
		},
		Max: y3d.Vec3{
			X: -999999.99,
			Y: -999999.99,
			Z: -999999.99,
		},
	}
	count := dataI.Len() / 4
	i := dataI.Bytes()
	indices := unsafe.Slice((*uint32)(unsafe.Pointer(unsafe.SliceData(i))), count)
	for _, i := range indices {
		vertex := getPos(int(i))
		box.Max.X = max(vertex.X, box.Max.X)
		box.Max.Y = max(vertex.Y, box.Max.Y)
		box.Max.Z = max(vertex.Z, box.Max.Z)

		box.Min.X = min(vertex.X, box.Min.X)
		box.Min.Y = min(vertex.Y, box.Min.Y)
		box.Min.Z = min(vertex.Z, box.Min.Z)
	}
	return Box{
		Local: box,
	}
}

func (b Box) RenderBoxMatrix() y3d.Mat4 {
	c := b.World.GetCenter()
	s := b.World.GetHalfSize()
	return y3d.Mat4{
		s.X, 0, 0, 0,
		0, s.Y, 0, 0,
		0, 0, s.Z, 0,
		c.X, c.Y, c.Z, 1.0,
	}
}
