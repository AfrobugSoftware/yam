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

//how to handle prefabs
