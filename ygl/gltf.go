package ygl

import (
	"errors"
	"fmt"
	"yam/y3d"
	"yam/yecs"

	"github.com/qmuntal/gltf"
)

func (g *Gl3) LoadAsset(filename string, w *yecs.World) error {
	doc, err := gltf.Open(filename)
	if err != nil {
		return err
	}
	fmt.Print(doc.Asset)
	if doc.Scene == nil {
		return errors.New("no default scene in asset")
	}
	scene := doc.Scenes[*doc.Scene]
	for _, n := range scene.Nodes {
		node := doc.Nodes[n]
		e := w.NewEntity()
		g.ProcessNode(node, doc, w, e, yecs.NullEntity)
	}
	return err
}

func (g *Gl3) ProcessNode(node *gltf.Node, doc *gltf.Document, w *yecs.World, e yecs.EntityId, parent yecs.EntityId) {
	if node.Camera != nil {
		//camera node

	}
	transfrom := yecs.NewTransfromation()
	if node.Mesh != nil {
		mesh := doc.Meshes[*node.Mesh]
		for _, p := range mesh.Primitives {

		}
		transfrom.Rotation = y3d.Quaternion{
			W: node.Rotation[0],
			X: node.Rotation[1],
			Y: node.Rotation[2],
			Z: node.Rotation[3],
		}
		transfrom.Position = y3d.Vec3{
			X: float32(node.Translation[0]),
			Y: float32(node.Translation[1]),
			Z: float32(node.Translation[2]),
		}
		(&transfrom).Recalulate()
	}
	children := make([]yecs.EntityId, 0, len(node.Children))
	for _, c := range node.Children {
		ne := w.NewEntity()
		children = append(children, ne)
		g.ProcessNode(doc.Nodes[c], doc, w, ne, e)
	}
	hc := yecs.Hierarchy{
		Parent:   parent,
		Children: children,
	}
	w.AddComponent(e, yecs.HierarchyComponent, hc)
	w.AddComponent(e, yecs.TransformComponent, transfrom)
}
