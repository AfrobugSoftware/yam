package ygame

import (
	"errors"
	"yam/y3d"
	"yam/ycore"
)

const (
	MAX_POLYGONS = 100_000
)

var (
	polyMeshComponent = ycore.RegisterComponent[Polymesh]()
)

type Polymesh struct {
	OType       int
	Polys       []Polygon
	DrawCommand ycore.DrawCommand
	Transform   ycore.Transform
	BoundingBox y3d.AABB
	SkinId      int
	IsDeleted   bool
	IsHidden    bool
	IsSelected  bool
	NumVerts    int
	NumIndices  int
}

func (pm *Polymesh) AddPolygon(p *Polygon) {
	p2 := &Polygon{}
	p2.Copy(p)
	p2.SkinId = pm.SkinId
	pm.Polys = append(pm.Polys, *p2)

	pm.NumIndices += len(p.Indices)
	pm.NumVerts += len(p.Vertices)

	box := pm.Polys[0].BoundingBox
	for i := range pm.Polys {
		ip := &pm.Polys[i]
		box.Max = y3d.Max(box.Max, ip.BoundingBox.Max)
		box.Min = y3d.Min(box.Min, ip.BoundingBox.Min)
	}
}

func (pm *Polymesh) RemovePolygon(i int) error {
	if i >= len(pm.Polys) {
		return errors.New("invalid polygon index")
	}
	pm.Polys[i].IsDeleted = true
	pm.NumIndices -= len(pm.Polys[i].Indices)
	pm.NumVerts -= len(pm.Polys[i].Vertices)
	box := pm.Polys[0].BoundingBox
	for i := range pm.Polys {
		ip := &pm.Polys[i]
		if !ip.IsDeleted {
			box.Max = y3d.Max(box.Max, ip.BoundingBox.Max)
			box.Min = y3d.Min(box.Min, ip.BoundingBox.Min)
		}
	}
	return nil
}

func (pm *Polymesh) GetSelected() []int {
	idx := make([]int, 0, 1000)
	for i := range pm.Polys {
		if pm.Polys[i].IsSelected {
			idx = append(idx, i)
		}
	}
	return idx
}

func (pm *Polymesh) Copy(pm2 *Polymesh) {
	*pm = *pm2
	clear(pm.Polys)
	copy(pm.Polys, pm2.Polys)
	box := pm.Polys[0].BoundingBox
	for i := range pm.Polys {
		ip := &pm.Polys[i]
		box.Max = y3d.Max(box.Max, ip.BoundingBox.Max)
		box.Min = y3d.Min(box.Min, ip.BoundingBox.Min)
	}
}

func (pm *Polymesh) Instersects(box y3d.AABB) bool {
	if pm.IsHidden {
		return false
	}
	if !pm.BoundingBox.Overlaps(box) {
		return false
	}
	for _, p := range pm.Polys {
		if p.IsHidden {
			continue
		}
		if p.BoundingBox.Overlaps(box) {
			return true
		}
	}
	return false
}

// if these things are pointers are they going to be stable references
func (pm *Polymesh) GetIntersectingPolys(box y3d.AABB, axis AXIS) ([]*Polygon, bool) {
	if pm.IsHidden {
		return nil, false
	}
	if !pm.BoundingBox.Overlaps(box) {
		return nil, false
	}
	pp := make([]*Polygon, 0)
	for i := range pm.Polys {
		if pm.Polys[i].IsHidden || pm.Polys[i].IsDeleted {
			continue
		}
		if pm.Polys[i].InBox(box, axis) {
			pp = append(pp, &pm.Polys[i])
		}
	}
	return pp, len(pp) > 0
}
func (pm *Polymesh) GetIntersectingVertices(box y3d.AABB, axis AXIS) ([]y3d.PVertex, bool) {
	if pm.IsHidden {
		return nil, false
	}
	if !pm.BoundingBox.Overlaps(box) {
		return nil, false
	}
	pp := make([]y3d.PVertex, 0, 1000)
	for i := range pm.Polys {
		if pm.Polys[i].IsHidden || pm.Polys[i].IsDeleted {
			continue
		}
		pv := pm.Polys[i].InsectingVertices(box, axis)
		pp = append(pp, pv...)
	}
	return pp, len(pp) > 0
}

func (pm *Polymesh) PickPoly(ray y3d.Ray, min float32) (*Polygon, float32, bool) {
	if pm.IsHidden {
		return nil, 0, false
	}
	b, _ := ray.IntersectsAABB(pm.BoundingBox, 0)
	var pp *Polygon
	if !b {
		for i := range pm.Polys {
			if pm.Polys[i].IsHidden || pm.Polys[i].IsDeleted {
				continue
			}
			b, h := pm.Polys[i].Picked(ray, 0)
			if b {
				if h < min {
					min = h
					pp = &pm.Polys[i]
				}
			}
		}
	}
	return pp, min, pp != nil
}

func (pm *Polymesh) Picked(ray y3d.Ray) (float32, bool) {
	if pm.IsHidden {
		return 0, false
	}
	b, _ := ray.IntersectsAABB(pm.BoundingBox, 0)
	if !b {
		return 0, false
	}
	_, h, b := pm.PickPoly(ray, 9999.9)
	return h, b
}

func (pm *Polymesh) Render(v *ycore.VertexCache) {
	pm.DrawCommand.BaseVertex = uint32(v.NumVertics)
	pm.DrawCommand.BaseInstance = uint32(v.NumOfInstances)
	for _, p := range pm.Polys {
		if p.IsHidden || p.IsDeleted {
			continue
		}
		p.WriteCache(v)
	}
	v.AddDrawCommand(pm.DrawCommand)
}

//how to handle prefabs
