package ycore

import (
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"os"
	"unsafe"
	"yam/y3d"

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
		err := ProcessNode(root, doc.Nodes[n], doc, r, data)
		if err != nil {
			return nil, err
		}
	}
	return root, nil
}

func ProcessNode(p *Node, node *gltf.Node, doc *gltf.Document, r *RenderManager, data map[int][]byte) error {
	ynode := NewNode(r, p, y3d.UnitAABB, NewTransform())
	if node.Mesh != nil {
		mesh := doc.Meshes[*node.Mesh]
		v, i := &bytes.Buffer{}, &bytes.Buffer{}
		vertexType := VP
		format := make([]VertexFormat, 0)
		for _, primitive := range mesh.Primitives {
			attrib := primitive.Attributes
			for _, i := range attrib {
				ass := doc.Accessors[i]
				vf := VertexFormat{
					ComponentSize:  int32(ass.Type.Components()),
					Type:           uint32(compMap[ass.ComponentType]),
					RelativeOffset: uint32(ass.ByteOffset), //how to calculate
				}
				format = append(format, vf)
				if ass.BufferView == nil {
					continue
				}
				bv := doc.BufferViews[*ass.BufferView]
				b, ok := data[bv.Buffer]
				if !ok {
					i, err := loadBufferURI(doc, ass)
					if err != nil {
						return err
					}
					data[bv.Buffer] = i
					b = i
				}
				//but we need to pack this buffer
				//I need to sleep
				v.Write(b[bv.ByteOffset : bv.ByteOffset+bv.ByteLength])
			}
			vertexType = GetVertexFormat(format, r)
			if vertexType == INVALID_VERTEX_FORMAT {
				return errors.New("unsuppored vertex format")
			}
			if primitive.Indices != nil {
				ass := doc.Accessors[*primitive.Indices]
				if ass.BufferView == nil {
					continue
				}
				bv := doc.BufferViews[*ass.BufferView]
				b, ok := data[bv.Buffer]
				if !ok {
					b, err := loadBufferURI(doc, ass)
					if err != nil {
						return err
					}
					data[bv.Buffer] = b
				}
				if ass.ComponentType == gltf.ComponentUshort {
					//convert to unsigned int
					b = b[bv.ByteOffset:bv.ByteLength]
					id := make([]uint32, len(b)/2)
					sb := unsafe.Slice((*uint16)(unsafe.Pointer(unsafe.SliceData(b))), len(b)/2)
					for _, i := range sb {
						id[i] = uint32(i)
					}
					b = unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(id))), len(id)*4)
					i.Write(b[bv.ByteOffset : bv.ByteOffset+(bv.ByteLength*2)])
				} else {
					i.Write(b[bv.ByteOffset : bv.ByteOffset+bv.ByteLength])
				}
			}
		}
		geo := NewGeometry(r, ynode, y3d.UnitAABB,
			NewTransform(),
			vertexType,
			v, i,
			r.VertextManager.CreateDrawCommand(i, 1),
			NO_SKINID, NO_STATICBUF)
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
	p.Add(ynode)

	for _, n := range node.Children {
		err := ProcessNode(ynode, doc.Nodes[n], doc, r, data)
		if err != nil {
			return err
		}
	}
	return nil
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

func GetVertexFormat(vf []VertexFormat, r *RenderManager) string {
	vm := r.VertextManager
	for vt, f := range vm.Formats {
		if len(f) == len(vf) {
			for i := range f {
				if f[i].ComponentSize == vf[i].ComponentSize &&
					f[i].RelativeOffset == vf[i].RelativeOffset &&
					f[i].Type == vf[i].Type {
					return vt
				}
			}
		}
	}
	return INVALID_VERTEX_FORMAT
}
