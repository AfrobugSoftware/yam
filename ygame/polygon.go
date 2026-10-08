package ygame

import (
	"slices"
	"yam/y3d"
	"yam/ycore"
)

var (
	PolygonComponent = ycore.RegisterComponent[Polygon]()
)

type Polygon struct {
	Vertices     []y3d.PVertex
	Indices      []uint32
	LineIndicies []uint32
	TexOff       y3d.Vec2
	TexRep       y3d.Vec2
	SkinId       int
	IsDeleted    bool
	IsHidden     bool
	IsSelected   bool
	BoundingBox  y3d.AABB
	OType        int
	DrawCommand  *ycore.DrawCommand
}

func (p2 *Polygon) Copy(p *Polygon) {
	p2.SetVertices(p.Vertices, p.Indices)
	p2.TexRep = p.TexRep
	p2.TexOff = p.TexOff
	p2.SkinId = p.SkinId
	p2.OType = LOB_POLYGON
}

func (p *Polygon) Translate(pos y3d.Vec3) {
	for i := range p.Vertices {
		p.Vertices[i].Pos = y3d.Add(p.Vertices[i].Pos, pos)
	}
}

func (p *Polygon) Rotate(origin y3d.Vec3, rot y3d.Quaternion) {
	for i := range p.Vertices {
		p.Vertices[i].Pos = y3d.Sub(p.Vertices[i].Pos, origin)
		p.Vertices[i].Pos = rot.Rotate(p.Vertices[i].Pos)
		p.Vertices[i].Pos = y3d.Add(p.Vertices[i].Pos, origin)
	}
}

func (p *Polygon) Mirror(origin, axis y3d.Vec3) {
	for i := range p.Vertices {
		p.Vertices[i].Pos = y3d.Sub(p.Vertices[i].Pos, origin)
		p.Vertices[i].Pos = y3d.Mul(p.Vertices[i].Pos, y3d.NegateVec3(axis))
		p.Vertices[i].Pos = y3d.Add(p.Vertices[i].Pos, origin)
	}
}

func (p *Polygon) SetVertices(verts []y3d.PVertex, indics []uint32) {
	p.Vertices = make([]y3d.PVertex, len(verts))
	copy(p.Vertices, verts)
	p.LineIndicies = make([]uint32, len(verts)*2)
	for i := 0; i <= len(verts); i += 2 {
		p.LineIndicies[(i * 2)] = uint32(i)
		p.LineIndicies[(i*2)+1] = uint32(i + 1)
		if i == len(verts)-1 {
			p.LineIndicies[(i*2)+1] = 0
		}
	}
	box := y3d.AABB{
		Min: p.Vertices[0].Pos,
		Max: p.Vertices[0].Pos,
	}
	for _, v := range verts {
		box.Max = y3d.Max(box.Max, v.Pos)
		box.Min = y3d.Min(box.Min, v.Pos)
	}
	p.BoundingBox = box
	p.Indices = make([]uint32, len(indics))
	copy(p.Indices, indics)
	for i := 0; i < len(p.Indices); i += 3 {
		a := p.Vertices[p.Indices[i]].Pos
		b := p.Vertices[p.Indices[i+1]].Pos
		c := p.Vertices[p.Indices[i+2]].Pos

		d := y3d.Sub(b, a)
		e := y3d.Sub(c, a)

		n := y3d.Cross(d, e)
		p.Vertices[p.Indices[i]].Norm = n
		p.Vertices[p.Indices[i+1]].Norm = n
		p.Vertices[p.Indices[i+2]].Norm = n
	}
}

func (p *Polygon) InsideOut() {
	slices.Reverse(p.Indices)
	for i := 0; i < len(p.Indices); i += 3 {
		a := p.Vertices[p.Indices[i]].Pos
		b := p.Vertices[p.Indices[i+1]].Pos
		c := p.Vertices[p.Indices[i+2]].Pos

		d := y3d.Sub(b, a)
		e := y3d.Sub(c, a)

		n := y3d.Cross(d, e)
		p.Vertices[p.Indices[i]].Norm = n
		p.Vertices[p.Indices[i+1]].Norm = n
		p.Vertices[p.Indices[i+2]].Norm = n
	}
}
func (p *Polygon) Picked(ray y3d.Ray, fl float32) (bool, float32) {
	b, _ := ray.IntersectsAABB(p.BoundingBox, fl)
	if b {
		for i := 0; i < len(p.Indices); i += 3 {
			i0 := p.Indices[i]
			i1 := p.Indices[i+1]
			i2 := p.Indices[i+2]

			return ray.IntersectsTriangle(p.Vertices[i0].Pos, p.Vertices[i1].Pos,
				p.Vertices[i2].Pos, false)
		}
	}
	return false, 0
}

func (p *Polygon) InBox(box y3d.AABB, axis AXIS) bool {
	for _, p := range p.Vertices {
		v := p.Pos
		switch axis {
		case X_AXIS:
			if v.Y > box.Min.Y &&
				v.Y < box.Max.Y &&
				v.Z > box.Min.Z &&
				v.Z < box.Max.Z {
				return true
			}
		case Y_AXIS:
			if v.X > box.Min.X &&
				v.X < box.Max.X &&
				v.Z > box.Min.Z &&
				v.Z < box.Max.Z {
				return true
			}
		case Z_AXIS:
			if v.Y > box.Min.Y &&
				v.Y < box.Max.Y &&
				v.X > box.Min.X &&
				v.X < box.Max.X {
				return true
			}
		default:
			return false
		}
	}
	return false
}

func (p *Polygon) CalculateTextureCoordinates(axis AXIS, box *y3d.AABB) {
	var abox y3d.AABB
	if box != nil {
		abox = *box
	} else {
		abox = p.BoundingBox
	}
	p.TexOff = y3d.Vec2{}
	p.TexRep = y3d.Vec2{X: 1.0, Y: 1.0}
	size := y3d.Sub(abox.Max, abox.Min)
	for i := range p.Vertices {
		switch axis {
		case X_AXIS:
			p.Vertices[i].Tc.X = (p.Vertices[i].Pos.Z - abox.Min.Z) / size.Z
			p.Vertices[i].Tc.Y = (p.Vertices[i].Pos.Y - abox.Min.Y) / size.Y
		case Y_AXIS:
			p.Vertices[i].Tc.X = (p.Vertices[i].Pos.X - abox.Min.X) / size.X
			p.Vertices[i].Tc.Y = (p.Vertices[i].Pos.Z - abox.Min.Z) / size.Z
		case Z_AXIS:
			p.Vertices[i].Tc.X = (p.Vertices[i].Pos.X - abox.Min.X) / size.X
			p.Vertices[i].Tc.Y = (p.Vertices[i].Pos.Y - abox.Min.Y) / size.Y
		}
	}
}

func (p *Polygon) WriteCache(v *ycore.VertexCache) error {
	return v.AddPolygon(p.Indices, p.Vertices)
}
