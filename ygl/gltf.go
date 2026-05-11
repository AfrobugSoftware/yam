package ygl

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"yam/y3d"
	"yam/yecs"

	"github.com/qmuntal/gltf"
)

var (
	bufferCache   = map[int][]byte{}
	nodeMap       = map[int]yecs.EntityId{}
	bufferCacheMu sync.Mutex
)

func LoadAsset(filename string, w *yecs.World) ([]yecs.EntityId, error) {
	bufferCacheMu.Lock()
	defer bufferCacheMu.Unlock()

	doc, err := gltf.Open(filename)
	if err != nil {
		return nil, err
	}
	fmt.Print(doc.Asset.Version)
	if doc.Scene == nil {
		return nil, errors.New("no default scene in asset")
	}
	scene := doc.Scenes[*doc.Scene]
	mesh := w.NewMesh()
	ents := make([]yecs.EntityId, 0)
	for _, n := range scene.Nodes {
		node := doc.Nodes[n]
		e := w.NewEntity()
		processNode(node, doc, w, mesh, e, yecs.NullEntity)
		ents = append(ents, e)
	}
	mesh.Setup()
	clear(bufferCache)
	return ents, err
}

func isValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func loadBufferURI(doc *gltf.Document, accessor *gltf.Accessor) ([]byte, error) {
	if accessor.BufferView == nil {
		return nil, nil //no buffer to load, is this an error ?
	}
	bv := doc.BufferViews[*accessor.BufferView]
	buffer := doc.Buffers[bv.Buffer]

	d, ok := bufferCache[bv.Buffer]
	if ok {
		return d[bv.ByteOffset : bv.ByteLength+bv.ByteOffset], nil
	}

	if buffer.IsEmbeddedResource() {
		s := strings.Split(buffer.URI, ",")
		if len(s) != 2 {
			return nil, errors.New("invalid embeded resource")
		}
		decoded, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			return nil, fmt.Errorf("invalid embeded resource: %v", err)

		}
		bufferCache[bv.Buffer] = decoded
		return decoded[bv.ByteOffset : bv.ByteLength+bv.ByteOffset], nil
	} else {
		if isValidURL(buffer.URI) {
			return nil, errors.New("cannot get data from endpoint, not supported")
		}
		file, err := os.Open(buffer.URI)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		decoded := make([]byte, bv.ByteLength)
		n, err := file.ReadAt(decoded, int64(bv.ByteOffset))
		if err != nil { //here io.EOF is an error
			return nil, err
		}
		if bv.ByteLength != n {
			return nil, errors.New("could not read the buffer data given in the buffer view")
		}
		bufferCache[bv.Buffer] = decoded
		return decoded[bv.ByteOffset : bv.ByteLength+bv.ByteOffset], nil
	}
}

func createFormart(accessor *gltf.Accessor) yecs.GltfDataFormat {
	df := yecs.GltfDataFormat{}
	df.ComponentType = accessor.ComponentType
	df.ValueCount = accessor.Count
	df.ValueType = accessor.Type
	return df
}

func loadAnimation(doc *gltf.Document, w *yecs.World) {

}

func processNode(node *gltf.Node, doc *gltf.Document, w *yecs.World, ms *yecs.Mesh, e yecs.EntityId, parent yecs.EntityId) {
	if node.Camera != nil {
		//camera node

	}
	transfrom := yecs.NewTransfromation()
	if node.Mesh != nil {
		mesh := doc.Meshes[*node.Mesh]
		meshEntries := make([]yecs.MeshEntry, 0)
		for _, p := range mesh.Primitives {
			meshEntry := yecs.MeshEntry{
				MeshId: ms.MeshId,
			}
			attrib := p.Attributes
			ap, ok := attrib[gltf.POSITION]
			if !ok {
				//at least the position is neccessary to load a mesh primitive
				continue
			}
			//mesh position attributes
			accessor := doc.Accessors[ap]
			buffer, err := loadBufferURI(doc, accessor)
			if err != nil {
				log.Println(err)
				continue
			}
			bv := doc.BufferViews[*accessor.BufferView]
			if bv.ByteStride == 0 {
				ms.Positions.Write(buffer[accessor.ByteOffset:])
			} else {
				//packdata
				for i := range accessor.Count {
					tcount := accessor.ComponentType.ByteSize() * accessor.Type.Components()
					start := accessor.ByteOffset + (i * bv.ByteStride)
					ms.Positions.Write(buffer[start : start+tcount])
				}
			}
			meshEntry.NumVertices = uint32(accessor.Count)
			meshEntry.BaseVertex = ms.NumVertices
			ms.NumVertices += uint32(accessor.Count)
			meshEntries = append(meshEntries, meshEntry)
			if accessor.Max != nil && accessor.Min != nil {
				box := yecs.Box{
					Local: y3d.AABB{
						Min: y3d.Vec3{
							X: float32(accessor.Min[0]),
							Y: float32(accessor.Min[1]),
							Z: float32(accessor.Min[2]),
						},
						Max: y3d.Vec3{
							X: float32(accessor.Max[0]),
							Y: float32(accessor.Max[1]),
							Z: float32(accessor.Max[2]),
						},
					},
				}
				w.AddComponent(e, yecs.BoxComponent, box)
			}
			an, ok := attrib[gltf.NORMAL]
			if ok {
				//mesh position attributes
				accessor := doc.Accessors[an]
				buffer, err := loadBufferURI(doc, accessor)
				if err != nil {
					log.Println(err)
					continue
				}
				bv := doc.BufferViews[*accessor.BufferView]
				if bv.ByteStride == 0 {
					ms.Normals.Write(buffer[accessor.ByteOffset:])
				} else {
					//packdata
					for i := range accessor.Count {
						tcount := accessor.ComponentType.ByteSize() * accessor.Type.Components()
						start := accessor.ByteOffset + (i * bv.ByteStride)
						ms.Normals.Write(buffer[start : start+tcount])
					}
				}
			}
			at, ok := attrib[gltf.TEXCOORD_0]
			if ok {
				accessor := doc.Accessors[at]
				buffer, err := loadBufferURI(doc, accessor)
				if err != nil {
					log.Println(err)
					continue
				}
				bv := doc.BufferViews[*accessor.BufferView]
				if bv.ByteStride == 0 {
					ms.TexCoords.Write(buffer[accessor.ByteOffset:])
				} else {
					//packdata
					for i := range accessor.Count {
						tcount := accessor.ComponentType.ByteSize() * accessor.Type.Components()
						start := accessor.ByteOffset + (i * bv.ByteStride)
						ms.TexCoords.Write(buffer[start : start+tcount])
					}
				}
			}
			if p.Indices != nil {
				accessor := doc.Accessors[*p.Indices]
				buffer, err := loadBufferURI(doc, accessor)
				if err != nil {
					continue
				}

				ms.Indices.Write(buffer[accessor.ByteOffset:])
				meshEntry.NumIndices = uint32(accessor.Count)
				ms.NumIndices += uint32(accessor.Count)
			}
			//material
		}
		w.AddComponent(e, yecs.MeshEntryComponent, meshEntries)
		transfrom.Rotation = y3d.Quaternion{
			X: node.Rotation[0],
			Y: node.Rotation[1],
			Z: node.Rotation[2],
			W: node.Rotation[3],
		}
		transfrom.Position = y3d.Vec3{
			X: float32(node.Translation[0]),
			Y: float32(node.Translation[1]),
			Z: float32(node.Translation[2]),
		}
		transfrom.Scale = y3d.Vec3{
			X: float32(node.Scale[0]),
			Y: float32(node.Scale[0]),
			Z: float32(node.Scale[0]),
		}
		(&transfrom).Recalulate()
	}
	children := make([]yecs.EntityId, 0, len(node.Children))
	for _, c := range node.Children {
		ne := w.NewEntity()
		children = append(children, ne)
		processNode(doc.Nodes[c], doc, w, ms, ne, e)
	}
	hc := yecs.Hierarchy{
		Parent:   parent,
		Children: children,
	}
	w.AddComponent(e, yecs.HierarchyComponent, hc)
	w.AddComponent(e, yecs.TransformComponent, transfrom)
}
