package y3d

import "math"

const (
	UP_NE = iota
	UP_NW
	UP_SE
	UP_SW
	LW_NE
	LW_NW
	LW_SE
	LW_SW
)

const (
	POLYS_PER_LEAF = 10
)

type Octree struct {
	Root        *Octree
	Children    [8]*Octree
	Polys       []Polygon
	BoundingBox AABB
	Parent      *Octree
	Position    int
}

func (o *Octree) UpdateBoundingBox(world Mat4) {
	box := AABB{
		Min: world.MulVec3(o.BoundingBox.Min),
		Max: world.MulVec3(o.BoundingBox.Max),
	}
	o.BoundingBox = box
	for i := range o.Children {
		if o.Children[i] != nil {
			o.Children[i].UpdateBoundingBox(world)
		}
	}
}
func (o *Octree) IsLeaf() bool {
	return o.Children[0] == nil
}

func (o *Octree) SetRelationship(r, p *Octree, pos int) {
	o.Root = r
	o.Parent = p
	o.Position = pos
}

func (o *Octree) CalculateBoundingBox() {
	if len(o.Polys) == 0 {
		o.BoundingBox = UnitAABB
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
	for _, p := range o.Polys {
		b := p.BoundingBox
		box.Max = Max(box.Max, b.Max)
		box.Min = Min(box.Min, b.Min)
	}
	o.BoundingBox = box
}

func NewOctree() *Octree {
	return &Octree{
		Polys:    make([]Polygon, 0),
		Position: -1,
	}
}

func (o *Octree) InitChild(pos int) {
	aabb := AABB{}
	center := o.BoundingBox.GetCenter()
	xmin, xcen, xmax := o.BoundingBox.Min.X, center.X, o.BoundingBox.Max.X
	ymin, ycen, ymax := o.BoundingBox.Min.Y, center.Y, o.BoundingBox.Max.Y
	zmin, zcen, zmax := o.BoundingBox.Min.Z, center.Z, o.BoundingBox.Max.Z

	switch pos {
	case UP_NW:
		aabb.Max = Vec3{X: xcen, Y: ymax, Z: zmax}
		aabb.Min = Vec3{X: xmin, Y: ycen, Z: zmin}
	case UP_NE:
		aabb.Max = Vec3{X: xmax, Y: ymax, Z: zmax}
		aabb.Min = Vec3{X: xcen, Y: ycen, Z: zcen}
	case UP_SW:
		aabb.Max = Vec3{X: xcen, Y: ymax, Z: zcen}
		aabb.Min = Vec3{X: xmin, Y: ycen, Z: zmin}
	case UP_SE:
		aabb.Max = Vec3{X: xmax, Y: ymax, Z: zcen}
		aabb.Min = Vec3{X: xcen, Y: ycen, Z: zmin}
	case LW_NW:
		aabb.Max = Vec3{X: xcen, Y: ycen, Z: zmax}
		aabb.Min = Vec3{X: xmin, Y: ymin, Z: zcen}
	case LW_NE:
		aabb.Max = Vec3{X: xmax, Y: ycen, Z: zmax}
		aabb.Min = Vec3{X: xcen, Y: ymin, Z: zcen}
	case LW_SW:
		aabb.Max = Vec3{X: xcen, Y: ycen, Z: zcen}
		aabb.Min = Vec3{X: xmin, Y: ymin, Z: zmin}
	case LW_SE:
		aabb.Max = Vec3{X: xmax, Y: ycen, Z: zcen}
		aabb.Min = Vec3{X: xcen, Y: ymin, Z: zmin}
	}
	o.Children[pos] = &Octree{
		Polys:       make([]Polygon, 0),
		BoundingBox: aabb,
		Parent:      o,
		Position:    pos,
		Root:        o.Root,
	}
}

func (o *Octree) ChopListToMe(polys []Polygon) {
	if len(polys) < 1 {
		return
	}
	clear(o.Polys)
	for i, p := range polys {
		if p.Flag == 0x01 {
			continue
		}
		choppedPoly := p
		class := choppedPoly.Cull(o.BoundingBox)
		switch class {
		case CULLED:
			continue
		case CLIPPED:
			choppedPoly = choppedPoly.ClipAABB(o.BoundingBox)
		case VISIBLE:
			polys[i].Flag = 0x01
		}
		o.Polys = append(o.Polys, choppedPoly)
	}
}

func (o *Octree) CreateChild() {
	if len(o.Polys) <= POLYS_PER_LEAF {
		return
	}
	for i := range 8 {
		o.InitChild(i)
		o.Children[i].ChopListToMe(o.Polys)
		o.Children[i].CreateChild()
	}
	clear(o.Polys)
	o.Polys = nil
}

func (o *Octree) BuildTree(polys []Polygon) {
	o.Root = o
	if len(polys) < 1 {
		return
	}
	copy(o.Polys, polys)
	o.CalculateBoundingBox()

	o.CreateChild()
}

func (o *Octree) TestCollision(b AABB) (bool, Plane) {
	if o != o.Root {
		if !o.BoundingBox.Overlaps(b) {
			return false, Plane{}
		}
	}
	if !o.IsLeaf() {
		for i := range 8 {
			if o.Children[i] != nil {
				collides, plane := o.Children[i].TestCollision(b)
				if collides {
					return true, plane
				}
			}
		}
		return false, Plane{}
	}
	for _, p := range o.Polys {
		if p.BoundingBox.Overlaps(b) {
			return true, p.Plane
		}
	}
	return false, Plane{}
}

func (o *Octree) TestCollisionWithRay(r Ray, fl float32) (bool, float32) {
	if o != o.Root {
		b, _ := r.IntersectsAABB(o.BoundingBox)
		if !b {
			return false, 0
		}
	}
	if !o.IsLeaf() {
		for i := range 8 {
			if o.Children[i] != nil {
				collides, t := o.Children[i].TestCollisionWithRay(r, fl)
				if collides && t < fl {
					return true, t
				}
			}
		}
		return false, 0
	}
	var fd float32 = 0
	var b bool
	for _, p := range o.Polys {
		h, b := p.IntersectsRay(r, false)
		if b {
			if h < fl {
				fl = h
				fd = h
				b = true
			}
		}
	}
	return b, fd
}

func (o *Octree) InsectDownwardsRay(origin Vec3, f float32) bool {
	if origin.Y < o.BoundingBox.Min.Y {
		return false
	}
	if origin.X < o.BoundingBox.Min.X || origin.X > o.BoundingBox.Max.X {
		return false
	}
	if origin.Z < o.BoundingBox.Min.Z || origin.Z > o.BoundingBox.Max.Z {
		return false
	}
	fd := float32(math.Abs(float64(o.BoundingBox.Max.Y - origin.Y)))
	if f < fd {
		return false
	}
	return false
}

func (o *Octree) GetFloor(origin Vec3) (bool, float32, Plane) {
	var pf float32
	var bhit bool
	var plane Plane
	if o == o.Root {
		pf = 99999.9
	}
	if !o.IsLeaf() {
		for i := range 8 {
			if o.Children[i].InsectDownwardsRay(origin, pf) {
				b, f, p := o.Children[i].GetFloor(origin)
				if b {
					pf = f
					plane = p
					bhit = b
				}
			}
		}
		return bhit, pf, plane
	}
	r := Ray{
		O: origin,
		D: Vec3{0, -1, 0},
	}
	for _, p := range o.Polys {
		box := p.BoundingBox
		if r.O.X < box.Min.X ||
			r.O.X > box.Max.X ||
			r.O.Y < box.Min.Y ||
			r.O.Z < box.Min.Z ||
			r.O.Z > box.Max.Z {
			continue
		}
		f, hit := p.IntersectsRay(r, false)
		if hit {
			pf = f
			plane = p.Plane
			bhit = hit
		}
	}
	return bhit, pf, plane
}
