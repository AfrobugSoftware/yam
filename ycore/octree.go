package ycore

import (
	"math"
	"slices"
	"sync"
	"yam/y3d"
)

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
	ENTITES_PER_LEAF = 10
)

type Octree struct {
	Root        *Octree
	World       *World
	Children    [8]*Octree
	Entites     []EntityId
	BoundingBox y3d.AABB
	Parent      *Octree
	Position    int
}

func (o *Octree) UpdateBoundingBox(world y3d.Mat4) {
	box := y3d.AABB{
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
	if len(o.Entites) == 0 {
		o.BoundingBox = y3d.UnitAABB
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
	for _, p := range o.Entites {
		b := o.World.GetComponent(p, BoundingBoxComponent).(Box).World
		box.Max = y3d.Max(box.Max, b.Max)
		box.Min = y3d.Min(box.Min, b.Min)
	}
	o.BoundingBox = box
}

func NewOctree(w *World) *Octree {
	return &Octree{
		World:    w,
		Entites:  make([]EntityId, 0),
		Position: -1,
	}
}

func (o *Octree) InitChild(pos int) {
	aabb := y3d.AABB{}
	center := o.BoundingBox.GetCenter()
	xmin, xcen, xmax := o.BoundingBox.Min.X, center.X, o.BoundingBox.Max.X
	ymin, ycen, ymax := o.BoundingBox.Min.Y, center.Y, o.BoundingBox.Max.Y
	zmin, zcen, zmax := o.BoundingBox.Min.Z, center.Z, o.BoundingBox.Max.Z

	switch pos {
	case UP_NW:
		aabb.Max = y3d.Vec3{X: xcen, Y: ymax, Z: zmax}
		aabb.Min = y3d.Vec3{X: xmin, Y: ycen, Z: zmin}
	case UP_NE:
		aabb.Max = y3d.Vec3{X: xmax, Y: ymax, Z: zmax}
		aabb.Min = y3d.Vec3{X: xcen, Y: ycen, Z: zcen}
	case UP_SW:
		aabb.Max = y3d.Vec3{X: xcen, Y: ymax, Z: zcen}
		aabb.Min = y3d.Vec3{X: xmin, Y: ycen, Z: zmin}
	case UP_SE:
		aabb.Max = y3d.Vec3{X: xmax, Y: ymax, Z: zcen}
		aabb.Min = y3d.Vec3{X: xcen, Y: ycen, Z: zmin}
	case LW_NW:
		aabb.Max = y3d.Vec3{X: xcen, Y: ycen, Z: zmax}
		aabb.Min = y3d.Vec3{X: xmin, Y: ymin, Z: zcen}
	case LW_NE:
		aabb.Max = y3d.Vec3{X: xmax, Y: ycen, Z: zmax}
		aabb.Min = y3d.Vec3{X: xcen, Y: ymin, Z: zcen}
	case LW_SW:
		aabb.Max = y3d.Vec3{X: xcen, Y: ycen, Z: zcen}
		aabb.Min = y3d.Vec3{X: xmin, Y: ymin, Z: zmin}
	case LW_SE:
		aabb.Max = y3d.Vec3{X: xmax, Y: ycen, Z: zcen}
		aabb.Min = y3d.Vec3{X: xcen, Y: ymin, Z: zmin}
	}
	if o.Children[pos] != nil {
		o.Children[pos].Position = pos
		o.Children[pos].BoundingBox = aabb
		o.Children[pos].Parent = o
		o.Children[pos].World = o.World
	} else {
		o.Children[pos] = &Octree{
			World:       o.World,
			Entites:     make([]EntityId, 0),
			BoundingBox: aabb,
			Parent:      o,
			Position:    pos,
			Root:        o.Root,
		}
	}

}

func (o *Octree) ChopListToMe(entites []EntityId) {
	if len(entites) == 0 {
		return
	}
	o.Entites = o.Entites[:0]
	entites = slices.DeleteFunc(entites, func(p EntityId) bool {
		if p == NullEntity {
			return false
		}
		pbox := o.World.GetComponent(p, BoundingBoxComponent).(Box).World
		if o.BoundingBox.Overlaps(pbox) {
			o.Entites = append(o.Entites, p)
			return true
		}
		return false
	})
}

func (o *Octree) CreateChild() {
	if len(o.Entites) <= ENTITES_PER_LEAF {
		return
	}
	var wg sync.WaitGroup
	for i := range 8 {
		o.InitChild(i)
		o.Children[i].ChopListToMe(o.Entites)
		wg.Go(o.Children[i].CreateChild)
	}
	wg.Wait()
	o.Entites = o.Entites[:0]
}

func (o *Octree) BuildTree(entties []EntityId) {
	o.Root = o
	if len(entties) == 0 {
		return
	}
	o.Entites = append(o.Entites, entties...)
	o.CalculateBoundingBox()
	o.CreateChild()
}

func (o *Octree) TestCollision(b y3d.AABB) (bool, EntityId) {
	if o != o.Root {
		if !o.BoundingBox.Overlaps(b) {
			return false, NullEntity
		}
	}
	if !o.IsLeaf() {
		for i := range 8 {
			if o.Children[i] != nil {
				collides, p := o.Children[i].TestCollision(b)
				if collides {
					return true, p
				}
			}
		}
		return false, NullEntity
	}
	for _, p := range o.Entites {
		box := o.World.GetComponent(p, BoundingBoxComponent).(Box).World
		if box.Overlaps(b) {
			return true, p
		}
	}
	return false, NullEntity
}

func (o *Octree) TestCollisionWithRay(r y3d.Ray, fl float32) (bool, EntityId, y3d.Vec3) {
	if o != o.Root {
		b, _ := r.IntersectsAABB(o.BoundingBox, fl)
		if !b {
			return false, NullEntity, y3d.Vec3{}
		}
	}
	if !o.IsLeaf() {
		for i := range 8 {
			if o.Children[i] != nil {
				collides, e, h := o.Children[i].TestCollisionWithRay(r, fl)
				if collides {
					return true, e, h
				}
			}
		}
		return false, NullEntity, y3d.Vec3{}
	}
	var br bool
	var hit y3d.Vec3
	var e EntityId
	for _, p := range o.Entites {
		pbox := o.World.GetComponent(p, BoundingBoxComponent).(Box).World
		b, h := r.IntersectsAABB(pbox, fl)
		dist := y3d.Sub(h, r.O).Length()
		if b {
			if dist < fl {
				fl = dist
				br = true
				e = p
				hit = h
			}
		}
	}
	return br, e, hit
}

func (o *Octree) InsectDownwardsRay(origin y3d.Vec3, f float32) bool {
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

func (o *Octree) GetFloor(origin y3d.Vec3, from EntityId) (bool, y3d.Vec3) {
	var pf float32
	var bhit bool
	var hit y3d.Vec3
	if o == o.Root {
		pf = 99999.9
	}
	if !o.IsLeaf() {
		for i := range 8 {
			if o.Children[i].InsectDownwardsRay(origin, pf) {
				b, h := o.Children[i].GetFloor(origin, from)
				if b {
					hit = h
					bhit = b
				}
			}
		}
		return bhit, hit
	}
	r := y3d.Ray{
		O: origin,
		D: y3d.Vec3{X: 0, Y: -1, Z: 0},
	}
	smallest := math.MaxFloat32
	for _, p := range o.Entites {
		if p == from {
			continue //remove self intersection
		}
		box := o.World.GetComponent(p, BoundingBoxComponent).(Box).World
		if r.O.X < box.Min.X ||
			r.O.X > box.Max.X ||
			r.O.Y < box.Min.Y ||
			r.O.Z < box.Min.Z ||
			r.O.Z > box.Max.Z {
			continue
		}
		b, h := r.IntersectsAABB(box, pf)
		dist := y3d.Sub(h, r.O).Length()
		if b {
			if dist < float32(smallest) {
				smallest = float64(dist)
				hit = h
				bhit = b
			}
		}
	}
	return bhit, hit
}
