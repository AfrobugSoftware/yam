package ycore

import (
	"encoding/gob"
	"yam/y3d"
)

type Node struct {
	Spatial
	Children []SpatialInterface
}

func NewNode(
	renderManager *RenderManager,
	parent SpatialInterface,
	boundingBox y3d.AABB,
	tranform *Transform,
) *Node {
	return &Node{
		Spatial: Spatial{
			RenderManager:    renderManager,
			Parent:           parent,
			LocalBoundingBox: boundingBox,
			Transform:        tranform,
		},
	}
}

func (n *Node) Write(e *gob.Encoder) error {
	return e.Encode(*n)
}

func (n *Node) Read(d *gob.Decoder) error {
	return d.Decode(n)
}

func (n *Node) Add(s SpatialInterface) {
	s.SetParent(n.Parent)
	n.Children = append(n.Children, s)

	//update world
	//updaate bound
	//propagate to the root if not root
}

func (n *Node) UpdateWorldBound() {
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
	for _, s := range n.Children {
		s.UpdateWorldBound()

		b := s.GetBoundingBox()
		box.Max = y3d.Max(box.Max, b.Max)
		box.Min = y3d.Min(box.Min, b.Min)
	}
	n.WorldBoundingBox = box
	// if n.Parent != nil {
	// 	n.Parent.UpdateWorldBound()
	// }
}

func (n *Node) UpdateWorldTransform() {
	n.Spatial.UpdateWorldTransform()
	for _, s := range n.Children {
		s.UpdateWorldTransform()
	}
}

func (n *Node) UpdateControllers(dt float32) {
	n.Spatial.UpdateControllers(dt)
	for _, s := range n.Children {
		s.UpdateControllers(dt)
	}
}

func (n *Node) Draw() {
	for _, s := range n.Children {
		s.Draw()
	}
}
