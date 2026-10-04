package y3d

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

func MakePolygonListFromVertexBuffer(v, i *bytes.Buffer, componentSize, relativeOffset, stride int) []Polygon {
	b := v.Bytes()
	getPos := func(i int) Vec3 {
		off := (int(stride) * i) + relativeOffset
		end := off + componentSize*int(unsafe.Sizeof(float32(0)))
		v := b[off:end]
		vp := unsafe.Slice((*float32)(unsafe.Pointer(unsafe.SliceData(v))), 3)
		return Vec3{
			X: vp[0],
			Y: vp[1],
			Z: vp[2],
		}
	}
	ib := i.Bytes()
	count := i.Len() / 4
	pl := make([]Polygon, count/3)
	indices := unsafe.Slice((*uint32)(unsafe.Pointer(&ib[0])), count)
	for ic := 0; ic < count-3; ic += 3 {
		i1, i2, i3 := indices[ic], indices[ic+1], indices[ic+2]
		pos1 := getPos(int(i1))
		pos2 := getPos(int(i2))
		pos3 := getPos(int(i3))

		p := NewPolygon([]Vec3{pos1, pos2, pos3}, []uint{0, 1, 2})
		pl = append(pl, p)
	}

	return pl
}

// for debug rendering
func MakeVertexBufferFromPolygon(polys []Polygon, v, i *bytes.Buffer) {
	base := uint32(0)
	for _, p := range polys {
		normal := p.Plane.N
		maxX := p.BoundingBox.Max.X - p.BoundingBox.Min.X
		maxY := p.BoundingBox.Max.Y - p.BoundingBox.Min.Y
		for _, pt := range p.Points {
			binary.Write(v, binary.NativeEndian, pt.ToSlice())
			binary.Write(v, binary.NativeEndian, normal.ToSlice())
			tx := [2]float32{
				pt.X - p.BoundingBox.Min.X/maxX,
				pt.Y - p.BoundingBox.Min.Y/maxY,
			}
			binary.Write(v, binary.NativeEndian, tx)
		}
		for _, in := range p.Indices {
			binary.Write(i, binary.NativeEndian, uint32(in)+base)
		}
		base += uint32(len(p.Indices))
	}
}
