package ycore

import (
	"math"
	"yam/y3d"
)

type BSPTree struct {
	World       *World
	BoundingBox y3d.AABB
	Plane       y3d.Plane
	Back        *BSPTree
	Front       *BSPTree
	Root        *BSPTree
	Parent      *BSPTree
	Children    []EntityId //if leaf it would have a bunch of this
	Spliter     EntityId
}

func (bs *BSPTree) SetRelationShip(r, d *BSPTree) {
	bs.Parent = d
	bs.Root = r
}

func (bs *BSPTree) IsLeaf() bool {
	return bs.Back == nil && bs.Front == nil
}

func (bs *BSPTree) CalculateBoundingBox() {
	if len(bs.Children) == 0 {
		bs.BoundingBox = y3d.UnitAABB
		return
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
	for _, p := range bs.Children {
		b := bs.World.GetComponent(p, BoundingBoxComponent).(Box).World
		box.Max = y3d.Max(box.Max, b.Max)
		box.Min = y3d.Min(box.Min, b.Min)
	}
	bs.BoundingBox = box
}
func (bs *BSPTree) FindBestSplitter() bool {
	var front, back, planer, splits, bestSplitter int
	var found bool
	iscore, bestScore := 1000000, 1000000
	var plane y3d.Plane
	for i := range bs.Children {
		front, back, splits = 0, 0, 0
		p1 := bs.Children[i]
		flag := bs.World.GetComponent(p1, FlagComponent).(uint32)
		pbox := bs.World.GetComponent(p1, BoundingBoxComponent).(Box).World
		if flag == 0x01 {
			continue
		}
		for j := range bs.Children {
			if i == j {
				continue
			}
			plane = pbox.GetPlanes()[0]
			jbox := bs.World.GetComponent(bs.Children[j], BoundingBoxComponent).(Box).World
			class := jbox.WhichSide(plane)
			switch class {
			case y3d.FRONT:
				front++
			case y3d.BACK:
				back++
			case y3d.PLANER:
				planer++
			default:
				splits++
			}
		}
		x := front - back
		if x < 0 {
			x = -x
		}
		iscore = int(x + (splits * 3))
		if iscore < bestScore {
			if front > 0 || back > 0 || splits > 0 {
				bestScore = iscore
				bestSplitter = i
				found = true
			}
		}
	}
	if !found {
		return false
	}
	bs.Spliter = bs.Children[bestSplitter]
	bs.World.SetComponent(bs.Spliter, FlagComponent, uint32(0x01))
	bs.Plane = plane
	return true
}

func (bs *BSPTree) CreateChild() {
	bs.CalculateBoundingBox()
	if !bs.FindBestSplitter() {
		return
	}
	if bs.Front == nil {
		bs.Front = &BSPTree{
			Children: make([]EntityId, len(bs.Children)),
		}
	}
	if bs.Back == nil {
		bs.Back = &BSPTree{
			Children: make([]EntityId, len(bs.Children)),
		}
	}
	bs.Front.SetRelationShip(bs.Root, bs)
	bs.Back.SetRelationShip(bs.Root, bs)

	for _, p := range bs.Children {
		if p == bs.Spliter {
			//always add the spliter to the back child
			//because we are using the  right plane of the AABB
			bs.Back.Children = append(bs.Back.Children, p)
			continue
		}
		pbox := bs.World.GetComponent(p, BoundingBoxComponent).(Box).World
		class := bs.Plane.ClassifyAABB(pbox)
		switch class {
		case y3d.FRONT:
			bs.Front.Children = append(bs.Front.Children, p)
		case y3d.BACK:
			bs.Back.Children = append(bs.Back.Children, p)
		case y3d.CLIPPED:
			bs.Front.Children = append(bs.Front.Children, p)
			bs.Back.Children = append(bs.Back.Children, p)
		}
	}
	bs.Children = bs.Children[:0]
	bs.Front.CreateChild()
	bs.Back.CreateChild()
}

func (bs *BSPTree) BuildTree(entities []EntityId) {
	bs.Root = bs
	bs.Parent = nil
	copy(bs.Children, entities)

	bs.CreateChild()
}

func (bs *BSPTree) TestCollision(r y3d.Ray, fl float32) (bool, y3d.Vec3) {
	if bs.IsLeaf() {
		for _, p := range bs.Children {
			pbox := bs.World.GetComponent(p, BoundingBoxComponent).(Box).World
			b, h := r.IntersectsAABB(pbox, fl)
			if b {
				return b, h
			}
		}
		return false, y3d.Vec3{}
	}
	class := bs.Plane.Classify(r.O)
	rf, rb, b := bs.Plane.ClipRay(r, fl)
	if b {
		switch class {
		case y3d.BACK:
			c, n := bs.Back.TestCollision(rb, fl)
			if c {
				return c, n
			} else {
				return bs.Front.TestCollision(rf, fl)
			}
		case y3d.FRONT:
			c, n := bs.Front.TestCollision(rf, fl)
			if c {
				return c, n
			} else {
				return bs.Back.TestCollision(rb, fl)
			}
		}
	} else {
		switch class {
		case y3d.BACK:
			return bs.Back.TestCollision(r, fl)
		case y3d.FRONT:
			return bs.Front.TestCollision(r, fl)
		}
	}
	return false, y3d.Vec3{}
}

func (bs *BSPTree) Clear() {
	bs.Children = bs.Children[:0]
	if bs.Front != nil {
		bs.Front.Clear()
	}
	if bs.Back != nil {
		bs.Back.Clear()
	}
}

func (bs *BSPTree) LineOfSight(pos, target y3d.Vec3) bool {
	r := y3d.Ray{
		O: pos,
		D: y3d.Normalize(y3d.Sub(target, pos)),
	}
	c, _ := bs.TestCollision(r, (y3d.Sub(target, pos)).Length())
	return !c
}

func (bs *BSPTree) SelectFromPickRay(r y3d.Ray, fl float32) (EntityId, bool) {
	smallest := math.MaxFloat32
	found := false
	ent := NullEntity
	if bs.IsLeaf() {
		for _, p := range bs.Children {
			pbox := bs.World.GetComponent(p, BoundingBoxComponent).(Box).World
			b, h := r.IntersectsAABB(pbox, fl)
			if b {
				dist := y3d.Sub(h, r.O).Length()
				if dist < float32(smallest) {
					smallest = float64(dist)
					ent = p
					found = true
				}
			}

		}
	} else {
		class := bs.Plane.Classify(r.O)
		rf, rb, b := bs.Plane.ClipRay(r, fl)
		if b {
			switch class {
			case y3d.BACK:
				e, b := bs.Back.SelectFromPickRay(rb, fl)
				if b {
					return e, b
				} else {
					return bs.Front.SelectFromPickRay(rf, fl)
				}
			case y3d.FRONT:
				e, b := bs.Front.SelectFromPickRay(rf, fl)
				if b {
					return e, b
				} else {
					return bs.Back.SelectFromPickRay(rb, fl)
				}
			}
		} else {
			switch class {
			case y3d.BACK:
				return bs.Back.SelectFromPickRay(r, fl)
			case y3d.FRONT:
				return bs.Front.SelectFromPickRay(r, fl)
			}
		}
	}
	return ent, found
}
