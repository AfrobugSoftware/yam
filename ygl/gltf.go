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
	"unsafe"
	"yam/y3d"
	"yam/yecs"

	"github.com/qmuntal/gltf"
)

var (
	bufferCache   = map[int][]byte{}
	nodeMap       = map[int]yecs.EntityId{}
	bufferCacheMu sync.Mutex
)

func (g *Gl3) LoadAsset(filename string, w *yecs.World) error {
	bufferCacheMu.Lock()
	defer bufferCacheMu.Unlock()

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
		g.processNode(node, doc, w, e, yecs.NullEntity)
	}

	clear(bufferCache)
	return err
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
		return d, nil
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
		return decoded, nil
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
		return decoded, nil
	}
}

func createFormart(accessor *gltf.Accessor) DataFormat {
	df := DataFormat{}
	df.Count = int32(accessor.Type.Components())
	df.ComponentType = ComponentType(accessor.ComponentType)

	return df
}

func (g *Gl3) loadAnimation(doc *gltf.Document, w *yecs.World) {

}

func (g *Gl3) processNode(node *gltf.Node, doc *gltf.Document, w *yecs.World, e yecs.EntityId, parent yecs.EntityId) {
	if node.Camera != nil {
		//camera node

	}
	transfrom := yecs.NewTransfromation()
	if node.Mesh != nil {
		mesh := doc.Meshes[*node.Mesh]
		meshSilo := map[string][]byte{}
		indicesData := make([]uint16, 0)
		for _, p := range mesh.Primitives {
			attrib := p.Attributes
			ap, ok := attrib[gltf.POSITION]
			if ok {
				//mesh position attributes
				accessor := doc.Accessors[ap]
				buffer, err := loadBufferURI(doc, accessor)
				if err != nil {
					log.Println(err)
					continue
				}
				meshSilo[gltf.POSITION] = buffer
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

				//how to handle sparse accessor
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
				meshSilo[gltf.NORMAL] = buffer
			}
			if p.Indices != nil {
				accessor := doc.Accessors[*p.Indices]
				buffer, err := loadBufferURI(doc, accessor)
				if err != nil {
					continue
				}
				bufAsUint16 := unsafe.Slice((*uint16)(unsafe.Pointer(&buffer[0])), len(buffer)/int(unsafe.Sizeof(uint16(0))))
				indicesData = append(indicesData, bufAsUint16...)
			}
		}

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
		g.processNode(doc.Nodes[c], doc, w, ne, e)
	}
	hc := yecs.Hierarchy{
		Parent:   parent,
		Children: children,
	}
	w.AddComponent(e, yecs.HierarchyComponent, hc)
	w.AddComponent(e, yecs.TransformComponent, transfrom)
}
