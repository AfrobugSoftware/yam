package ycore

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"
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

type vertexAttrib struct {
	src    []byte
	size   int
	stride int
	count  int
}

var attribOrder = []string{
	"POSITION",
	"NORMAL",
	"TEXCOORD_0",
	"TEXCOORD_1",
	"COLOR_0",
	"TANGENT",
	"JOINTS_0",
	"WEIGHTS_0",
}

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

func attributeNames[V any](attrs map[string]V) []string {
	rank := func(name string) int {
		for k, n := range attribOrder {
			if n == name {
				return k
			}
		}
		return len(attribOrder)
	}
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Slice(names, func(a, b int) bool {
		ra, rb := rank(names[a]), rank(names[b])
		if ra != rb {
			return ra < rb
		}
		return names[a] < names[b]
	})
	return names
}

func align4(n int) int { return (n + 3) &^ 3 }

func makeVBuffer(v *bytes.Buffer, format []VertexFormat, attribs []vertexAttrib) (int, int, error) {
	if len(format) != len(attribs) {
		return 0, 0, errors.New("format list must match attribute list")
	}
	if len(attribs) == 0 {
		return 0, 0, nil
	}
	count := attribs[0].count
	vertexSize := 0
	for k, a := range attribs {
		if a.count != count {
			return 0, 0, fmt.Errorf("attribute %d has %d elements, expected %d", k, a.count, count)
		}
		if a.src != nil && count > 0 {
			if need := (count-1)*a.stride + a.size; len(a.src) < need {
				return 0, 0, fmt.Errorf("attribute %d: buffer too short (%d < %d)", k, len(a.src), need)
			}
		}
		format[k].RelativeOffset = uint32(vertexSize)
		vertexSize += align4(a.size) //GL wants 4-byte aligned attributes
	}
	out := make([]byte, count*vertexSize)
	for k, a := range attribs {
		if a.src == nil {
			continue
		}
		off := int(format[k].RelativeOffset)
		for n := range count {
			dst := n*vertexSize + off
			src := n * a.stride
			copy(out[dst:dst+a.size], a.src[src:src+a.size])
		}
	}
	v.Write(out)
	return count, vertexSize, nil
}

func appendIndices(ib *bytes.Buffer, ct gltf.ComponentType, a vertexAttrib, base uint32) error {
	if a.src == nil || len(a.src) < a.count*a.size {
		return errors.New("index accessor has no data")
	}
	var tmp [4]byte
	for n := 0; n < a.count; n++ {
		e := a.src[n*a.size:]
		var idx uint32
		switch ct {
		case gltf.ComponentUbyte:
			idx = uint32(e[0])
		case gltf.ComponentUshort:
			idx = uint32(binary.LittleEndian.Uint16(e))
		case gltf.ComponentUint:
			idx = binary.LittleEndian.Uint32(e)
		default:
			return errors.New("unsupported index component type")
		}
		binary.LittleEndian.PutUint32(tmp[:], idx+base)
		ib.Write(tmp[:])
	}
	return nil
}

func bufferData(doc *gltf.Document, acc *gltf.Accessor, idx int, data map[int][]byte) ([]byte, error) {
	if b, ok := data[idx]; ok {
		return b, nil
	}
	b, err := loadBufferURI(doc, acc)
	if err != nil {
		return nil, err
	}
	data[idx] = b
	return b, nil
}

func sameFormat(a, b []VertexFormat) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if a[k].ComponentSize != b[k].ComponentSize || a[k].Type != b[k].Type ||
			a[k].RelativeOffset != b[k].RelativeOffset || a[k].Normalized != b[k].Normalized {
			return false
		}
	}
	return true
}

func accessorAttrib(doc *gltf.Document, acc *gltf.Accessor, data map[int][]byte) (vertexAttrib, error) {
	size := int(acc.ComponentType.ByteSize()) * int(acc.Type.Components())
	a := vertexAttrib{size: size, stride: size, count: int(acc.Count)}
	if acc.BufferView == nil {
		return a, nil
	}
	bv := doc.BufferViews[*acc.BufferView]
	buf, err := bufferData(doc, acc, int(bv.Buffer), data)
	if err != nil {
		return a, err
	}
	if bv.ByteStride != 0 {
		a.stride = int(bv.ByteStride)
	}
	start := int(bv.ByteOffset) + int(acc.ByteOffset)
	end := int(bv.ByteOffset) + int(bv.ByteLength)
	if start > end || end > len(buf) {
		return a, errors.New("accessor out of buffer range")
	}
	a.src = buf[start:end]
	return a, nil
}

func addNodeSkin(p *Node, skin *gltf.Skin, node *gltf.Node, doc *gltf.Document, r *RenderManager) error {
	return nil
}

func ProcessNode(p *Node, node *gltf.Node, doc *gltf.Document, r *RenderManager, data map[int][]byte) error {
	ynode := NewNode(r, p, y3d.UnitAABB, NewTransform())
	if node.Mesh != nil {
		mesh := doc.Meshes[*node.Mesh]
		v, i := &bytes.Buffer{}, &bytes.Buffer{}
		vertexType := ""
		var format []VertexFormat
		var base uint32
		var vertexSize int
		for _, primitive := range mesh.Primitives {
			names := attributeNames(primitive.Attributes)
			pf := make([]VertexFormat, 0, len(names))
			attribs := make([]vertexAttrib, 0, len(names))
			for _, name := range names {
				acc := doc.Accessors[primitive.Attributes[name]]
				a, err := accessorAttrib(doc, acc, data)
				if err != nil {
					return fmt.Errorf("attribute %s: %w", name, err)
				}
				pf = append(pf, VertexFormat{
					ComponentSize: int32(acc.Type.Components()),
					Type:          uint32(compMap[acc.ComponentType]),
					Normalized:    acc.Normalized,
				})
				attribs = append(attribs, a)
				vertexType = strings.Join([]string{vertexType, name}, ".")
			}
			count, vs, err := makeVBuffer(v, pf, attribs)
			vertexSize = vs
			if err != nil {
				return err
			}
			if format == nil {
				format = pf
			} else if !sameFormat(format, pf) {
				return errors.New("primitives of one mesh must share a vertex format")
			}
			if primitive.Indices != nil {
				acc := doc.Accessors[*primitive.Indices]
				a, err := accessorAttrib(doc, acc, data)
				if err != nil {
					return fmt.Errorf("indices: %w", err)
				}
				if err := appendIndices(i, acc.ComponentType, a, base); err != nil {
					return err
				}
			} else {
				var tmp [4]byte
				for n := range count {
					binary.LittleEndian.PutUint32(tmp[:], base+uint32(n))
					i.Write(tmp[:])
				}
			}
			base += uint32(count)
		}
		tempVertexType := GetVertexFormat(format, r)
		if tempVertexType == INVALID_VERTEX_FORMAT {
			r.VertextManager.CreateVertexCache(vertexType, NO_SKINID, vertexSize, format)
		} else {
			vertexType = tempVertexType
		}
		geo := NewGeometry(r, ynode, y3d.UnitAABB,
			NewTransform(),
			vertexType,
			v, i,
			NO_SKINID,
			NO_STATICBUF,
			nil)
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
	if len(buffer.Data) > 0 {
		return buffer.Data, nil
	} else if buffer.IsEmbeddedResource() {
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
					f[i].Type == vf[i].Type &&
					f[i].Normalized == vf[i].Normalized {
					return vt
				}
			}
		}
	}
	return INVALID_VERTEX_FORMAT
}
