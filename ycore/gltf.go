package ycore

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"os"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/qmuntal/gltf"
)

const (
	mimetypeApplicationOctet = "data:application/octet-stream;base64"
	mimetypeImagePNG         = "data:image/png;base64"
	mimetypeImageJPG         = "data:image/jpeg;base64"
)

var (
	compMap = map[gltf.ComponentType]uint16{
		gltf.ComponentByte:   5120,
		gltf.ComponentUbyte:  5121,
		gltf.ComponentShort:  5122,
		gltf.ComponentUshort: 5123,
		gltf.ComponentUint:   5125,
		gltf.ComponentFloat:  5126,
	}
)

func LoadGLTF(filename string, r *RenderManager) (*Node, error) {
	doc, err := gltf.Open(filename)
	if err != nil {
		return nil, err
	}
	if doc.Scene == nil {
		return nil, errors.New("no scene in docuemt")
	}
	if doc.Asset.Version != "2.0" {
		return nil, errors.New("cannot parse gltf version")
	}
	scene := doc.Scenes[*doc.Scene]
	data := make(map[int][]byte)
	root := NewNode(r, nil, y3d.UnitAABB, NewTransform())
	for _, n := range scene.Nodes {
		ynode := NewNode(r, root, y3d.UnitAABB, NewTransform())
		node := doc.Nodes[n]
		if node.Mesh != nil {
			mesh := doc.Meshes[*node.Mesh]
			var v, i *bytes.Buffer
			vertexType := VP
			var idxType uint32 = gl.UNSIGNED_INT
			format := make([]VertexFormat, 0)
			for _, primitive := range mesh.Primitives {
				attrib := primitive.Attributes
				pos, ok := attrib[gltf.POSITION]
				if ok {
					ass := doc.Accessors[pos]
					if ass.BufferView == nil {
						continue
					}
					vf := VertexFormat{
						ComponentSize:  int32(ass.Type.Components()),
						Type:           uint32(compMap[ass.ComponentType]),
						RelativeOffset: uint32(ass.ByteOffset), //how to calculate
					}
					format = append(format, vf)
					bv := doc.BufferViews[*ass.BufferView]
					b, ok := data[bv.Buffer]
					if !ok {
						b, err = loadBufferURI(doc, ass)
						if err != nil {
							return nil, err
						}
						data[bv.Buffer] = b
					}

				}

				if primitive.Indices != nil {
					ass := doc.Accessors[*primitive.Indices]
					if ass.BufferView == nil {
						continue
					}
					bv := doc.BufferViews[*ass.BufferView]
					b, ok := data[bv.Buffer]
					if !ok {
						b, err = loadBufferURI(doc, ass)
						if err != nil {
							return nil, err
						}
						data[bv.Buffer] = b
					}
					idxType = uint32(compMap[ass.ComponentType])
					i.Write(b[bv.ByteOffset:bv.ByteLength])
				}
			}
			geo := NewGeometry(r, ynode, y3d.UnitAABB,
				NewTransform(),
				vertexType,
				v, i,
				DrawCommand{},
				-1, -1)
			ynode.Add(geo)
		}
		if node.Skin != nil {
			//skin := doc.Skins[*node.Skin]
			//for animation
		}
		if node.Camera != nil {
			//cam := doc.Cameras[*node.Camera]
			//set camera stage
			//how
		}
		root.Add(ynode)
	}
	return root, nil
}

func ProcessNode(node *gltf.Node, doc *gltf.Document, r *RenderManager, data map[int][]byte) {

}

func isValidURL(str string) bool {
	u, err := url.ParseRequestURI(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// reads the entire buffer
func loadBufferURI(doc *gltf.Document, accessor *gltf.Accessor) ([]byte, error) {
	if accessor.BufferView == nil {
		return nil, nil //no buffer to load, is this an error ?
	}
	bv := doc.BufferViews[*accessor.BufferView]
	buffer := doc.Buffers[bv.Buffer]
	if buffer.IsEmbeddedResource() {
		startPos := len(mimetypeApplicationOctet) + 1
		if len(buffer.URI) < startPos {
			return nil, errors.New("gltf: Invalid base64 content")
		}
		sl, err := base64.StdEncoding.DecodeString(buffer.URI[startPos:])
		if len(sl) == 0 || err != nil {
			return nil, err
		}
		return sl, nil
	} else {
		if isValidURL(buffer.URI) {
			return nil, errors.New("cannot get data from endpoint, not supported")
		}
		file, err := os.Open(buffer.URI)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		decoded, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		return decoded, nil
	}
}
