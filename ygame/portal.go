package ygame

import (
	"yam/y3d"
	"yam/ycore"
)

type Portal struct {
	vertices     []y3d.PVertex
	indices      []uint32
	lineIndicies []uint32
	Color        [4]uint8
	IsHidden     bool
	IsSelected   bool
	IsVisible    bool
	Id           int
	BrotherId    int
	Brother      *Portal
	BoundingBox  y3d.AABB
	Transform    ycore.Transform
	SkinId       int
	RenderFlag   ycore.PerObjectRendeState
	DrawCommand  ycore.DrawCommand
}

func NewPortal(verts []y3d.PVertex, indices []uint32) Portal {
	p := Portal{}
	p.vertices = make([]y3d.PVertex, len(verts))
	copy(p.vertices, verts)
	p.lineIndicies = make([]uint32, len(verts)*2)
	for i := 0; i <= len(verts); i += 2 {
		p.lineIndicies[(i * 2)] = uint32(i)
		p.lineIndicies[(i*2)+1] = uint32(i + 1)
		if i == len(verts)-1 {
			p.lineIndicies[(i*2)+1] = 0
		}
	}
	box := y3d.AABB{
		Min: p.vertices[0].Pos,
		Max: p.vertices[0].Pos,
	}
	for _, v := range verts {
		box.Max = y3d.Max(box.Max, v.Pos)
		box.Min = y3d.Min(box.Min, v.Pos)
	}
	p.BoundingBox = box
	p.indices = make([]uint32, len(indices))
	copy(p.indices, indices)
	for i := 0; i < len(p.indices); i += 3 {
		a := p.vertices[p.indices[i]].Pos
		b := p.vertices[p.indices[i+1]].Pos
		c := p.vertices[p.indices[i+2]].Pos

		d := y3d.Sub(b, a)
		e := y3d.Sub(c, a)

		n := y3d.Cross(d, e)
		p.vertices[p.indices[i]].Norm = n
		p.vertices[p.indices[i+1]].Norm = n
		p.vertices[p.indices[i+2]].Norm = n
		p.vertices[p.indices[i]].Color = p.Color
		p.vertices[p.indices[i+1]].Color = p.Color
		p.vertices[p.indices[i+2]].Color = p.Color
	}
	return p
}

func (p *Portal) Render(vm *ycore.VertexCacheManager) {
	v, err := vm.GetCache(ycore.VPNTC, p.SkinId)
	if err != nil {
		return
	}
	p.DrawCommand.BaseVertex = uint32(v.NumVertics)
	p.DrawCommand.Count = uint32(len(p.indices))
	err = v.AddPolygon(p.indices, p.vertices)
	if err != nil {
		return
	}
	v.AddInstances(&p.DrawCommand)
	v.AddRenderState(int(p.DrawCommand.BaseInstance), p.RenderFlag)
	v.LoadMatrix(int(p.DrawCommand.BaseInstance),
		p.Transform.World)
	v.AddDrawCommand(p.DrawCommand)
}

func (p *Portal) SetColor(color [4]uint8) {
	p.Color = color
	for i := 0; i < len(p.indices); i += 3 {

		p.vertices[p.indices[i]].Color = p.Color
		p.vertices[p.indices[i+1]].Color = p.Color
		p.vertices[p.indices[i+2]].Color = p.Color
	}
}

func (p *Portal) Picked(ray y3d.Ray, fl float32) (bool, float32) {
	b, _ := ray.IntersectsAABB(p.BoundingBox, fl)
	if b {
		for i := 0; i < len(p.indices); i += 3 {
			i0 := p.indices[i]
			i1 := p.indices[i+1]
			i2 := p.indices[i+2]

			return ray.IntersectsTriangle(p.vertices[i0].Pos, p.vertices[i1].Pos,
				p.vertices[i2].Pos, false)
		}
	}
	return false, 0
}

func (p *Portal) SetBrother(bro *Portal) {
	if bro == nil {
		p.BrotherId = 0
		p.Brother = nil
		return
	}
	if p.Id == bro.Id {
		return
	}
	p.BrotherId = bro.Id
	bro.BrotherId = p.Id
	p.Brother = bro
}
