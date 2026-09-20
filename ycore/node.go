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

func (n *Node) Draw() {
	for _, s := range n.Children {
		s.Draw()
	}
}
