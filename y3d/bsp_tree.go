package y3d

type BSPTree struct {
	BoundingBox AABB
	Plane       Plane
	Back        *BSPTree
	Front       *BSPTree
	Root        *BSPTree
	Parent      *BSPTree
	Polys       []Polygon //if leaf

}

func (bs *BSPTree) SetRelationShip(r, d *BSPTree) {
	bs.Parent = d
	bs.Root = r
}

func (bs *BSPTree) IsLeaf() bool {
	return bs.Back == nil && bs.Front == nil
}

func (bs *BSPTree) CalculateBoundingBox() {
	if len(bs.Polys) == 0 {
		bs.BoundingBox = UnitAABB
		return
	}
	box := AABB{
		Min: Vec3{
			X: 999999.99,
			Y: 999999.99,
			Z: 999999.99,
		},
		Max: Vec3{
			X: -999999.99,
			Y: -999999.99,
			Z: -999999.99,
		},
	}
	for _, p := range bs.Polys {
		b := p.BoundingBox
		box.Max = Max(box.Max, b.Max)
		box.Min = Min(box.Min, b.Min)
	}
	bs.BoundingBox = box
}
func (bs *BSPTree) FindBestSplitter() bool {
	var front, back, planer, splits, bestSplitter int
	var found bool
	iscore, bestScore := 1000000, 1000000
	for i := range bs.Polys {
		front, back, splits = 0, 0, 0
		p1 := bs.Polys[i]
		if p1.Flag == 0x01 {
			continue
		}
		for j := range bs.Polys {
			if i == j {
				continue
			}
			class := p1.Plane.ClassifyPolygon(bs.Polys[j])
			switch class {
			case FRONT:
				front++
			case BACK:
				back++
			case PLANER:
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
	bs.Polys[bestSplitter].Flag = 0x01
	bs.Plane = bs.Polys[bestSplitter].Plane
	return true
}

func (bs *BSPTree) CreateChild() {
	bs.CalculateBoundingBox()
	if !bs.FindBestSplitter() {
		return
	}
	bs.Front = &BSPTree{
		Polys: make([]Polygon, len(bs.Polys)),
	}
	bs.Back = &BSPTree{
		Polys: make([]Polygon, len(bs.Polys)),
	}
	bs.Front.SetRelationShip(bs.Root, bs)
	bs.Back.SetRelationShip(bs.Root, bs)

	for _, p := range bs.Polys {
		class := bs.Plane.ClassifyPolygon(p)
		switch class {
		case FRONT:
			bs.Front.Polys = append(bs.Front.Polys, p)
		case BACK:
			bs.Back.Polys = append(bs.Back.Polys, p)
		case CLIPPED:
			f, b := p.Clip(bs.Plane)
			bs.Front.Polys = append(bs.Front.Polys, f)
			bs.Back.Polys = append(bs.Back.Polys, b)
		case PLANER:
			dot := Dot(bs.Plane.N, p.Plane.N)
			if dot > 0.0 {
				bs.Front.Polys = append(bs.Front.Polys, p)
			} else {
				bs.Back.Polys = append(bs.Back.Polys, p)
			}
		}
	}
	clear(bs.Polys)
	bs.Polys = nil
	bs.Front.CreateChild()
	bs.Back.CreateChild()
}

func (bs *BSPTree) BuildTree(poly []Polygon) {
	bs.Root = bs
	bs.Parent = nil
	copy(bs.Polys, poly)

	bs.CreateChild()
}

func (bs *BSPTree) TransverseTreeFtB(polys []Polygon, pos Vec3, frustum []Plane) {
	if bs.BoundingBox.Cull(frustum) == CULLED {
		return
	}
	if bs.IsLeaf() {
		polys = append(polys, bs.Polys...)
	} else {
		class := bs.Plane.Classify(pos)
		switch class {
		case BACK:
			bs.Back.TransverseTreeFtB(polys, pos, frustum)
			bs.Front.TransverseTreeFtB(polys, pos, frustum)
		default:
			bs.Front.TransverseTreeFtB(polys, pos, frustum)
			bs.Back.TransverseTreeFtB(polys, pos, frustum)
		}
	}
}
func (bs *BSPTree) TransverseTreeBtF(polys []Polygon, pos Vec3, frustum []Plane) {
	if bs.BoundingBox.Cull(frustum) == CULLED {
		return
	}
	if bs.IsLeaf() {
		polys = append(polys, bs.Polys...)
	} else {
		class := bs.Plane.Classify(pos)
		switch class {
		case BACK:
			bs.Front.TransverseTreeBtF(polys, pos, frustum)
			bs.Back.TransverseTreeBtF(polys, pos, frustum)
		default:
			bs.Back.TransverseTreeBtF(polys, pos, frustum)
			bs.Front.TransverseTreeBtF(polys, pos, frustum)
		}
	}
}

func (bs *BSPTree) TestCollision(r Ray, fl float32) (bool, float32, Vec3) {
	if bs.IsLeaf() {
		for _, p := range bs.Polys {
			t, b := p.IntersectsRay(r, false)
			if b {
				return b, t, p.Plane.N
			}
		}
		return false, 0, ZEROV
	}
	class := bs.Plane.Classify(r.O)
	rf, rb, b := bs.Plane.ClipRay(r, fl)
	if b {
		switch class {
		case BACK:
			c, t, n := bs.Back.TestCollision(rb, fl)
			if c {
				return c, t, n
			} else {
				return bs.Front.TestCollision(rf, fl)
			}
		case FRONT:
			c, t, n := bs.Front.TestCollision(rf, fl)
			if c {
				return c, t, n
			} else {
				return bs.Back.TestCollision(rb, fl)
			}
		}
	} else {
		switch class {
		case BACK:
			return bs.Back.TestCollision(r, fl)
		case FRONT:
			return bs.Front.TestCollision(r, fl)
		}
	}
	return false, 0, ZEROV
}

func (bs *BSPTree) Clear() {
	clear(bs.Polys)
	bs.Polys = nil
	if bs.Front != nil {
		bs.Front.Clear()
	}
	if bs.Back != nil {
		bs.Back.Clear()
	}
}

func (bs *BSPTree) LineOfSight(pos, target Vec3) bool {
	r := Ray{
		O: pos,
		D: Normalize(Sub(target, pos)),
	}
	c, _, _ := bs.TestCollision(r, (Sub(target, pos)).Length())
	return !c
}

func (bs *BSPTree) SelectPolyFromPickRay(r Ray, fl float32) (Polygon, bool) {
	if bs.IsLeaf() {
		for _, p := range bs.Polys {
			if p.ContainsPoint(r.O) {
				return p, true
			}
		}
	}
	return Polygon{}, false
}
