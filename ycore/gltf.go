package ycore

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"net/url"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"
	"unsafe"
	"yam/y3d"
	"yam/ywd"

	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

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
	minFilter = map[gltf.MinFilter]uint16{
		gltf.MinUndefined:            gl.NEAREST,
		gltf.MinNearest:              9728,
		gltf.MinLinear:               9729,
		gltf.MinNearestMipMapNearest: 9984,
		gltf.MinLinearMipMapNearest:  9985,
		gltf.MinNearestMipMapLinear:  9986,
		gltf.MinLinearMipMapLinear:   9987,
	}
	magFilter = map[gltf.MagFilter]uint16{
		gltf.MagUndefined: gl.NEAREST,
		gltf.MagNearest:   9728,
		gltf.MagLinear:    9729,
	}
	wrapMode = map[gltf.WrappingMode]uint16{
		gltf.WrapClampToEdge:    33071,
		gltf.WrapMirroredRepeat: 33648,
		gltf.WrapRepeat:         10497,
	}
)

type childNode struct {
	parent int
	node   int
}

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

func LoadGLTF(filename string, r *RenderManager, world *ywd.World) error {
	doc, err := gltf.Open(filename)
	if err != nil {
		return err
	}
	if doc.Scene == nil {
		return errors.New("no scene in docuemt")
	}
	if doc.Asset.Version != "2.0" {
		return errors.New("cannot parse gltf version")
	}
	scene := doc.Scenes[*doc.Scene]
	data := make(map[int][]byte)
	skin := -1
	for _, n := range scene.Nodes {
		err := ProcessNode(world, doc.Nodes[n], doc, r, skin, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadTexture(doc *gltf.Document, texIndx int, skinManager *TextureManager) (uint32, error) {
	t := doc.Textures[texIndx]
	if t == nil {
		return 0, errors.New("no texture in doc")
	}
	sampler := doc.Samplers[*t.Sampler]
	img := doc.Images[*t.Source]
	if img.URI != "" {
		return skinManager.AddTexture(img.URI, int32(sampler.MinFilter),
			int32(sampler.MagFilter),
			int32(sampler.WrapS),
			int32(sampler.WrapT))
	} else {
		var texId uint32
		gl.CreateTextures(gl.TEXTURE_2D, 1, &texId)
		gl.TextureParameteri(texId, gl.TEXTURE_WRAP_S, int32(wrapMode[sampler.WrapS]))
		gl.TextureParameteri(texId, gl.TEXTURE_WRAP_T, int32(wrapMode[sampler.WrapT]))
		gl.TextureParameteri(texId, gl.TEXTURE_MIN_FILTER, int32(minFilter[sampler.MinFilter]))
		gl.TextureParameteri(texId, gl.TEXTURE_MAG_FILTER, int32(magFilter[sampler.MagFilter]))

		err := CheckError()
		if err != nil {
			return 0, err
		}

		bv := doc.BufferViews[*img.BufferView]
		buffer := doc.Buffers[bv.Buffer]
		start := int(bv.ByteOffset)
		end := int(bv.ByteOffset) + int(bv.ByteLength)
		data := buffer.Data[start:end]

		imgReader := bytes.NewReader(data)
		imgData, _, err := image.Decode(imgReader)
		if err != nil {
			return 0, err
		}
		bounds := imgData.Bounds()
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, imgData, bounds.Min, draw.Src)
		levels := mipLevels(
			int32(minFilter[sampler.MinFilter]),
			int32(bounds.Max.X),
			int32(bounds.Max.Y),
		)
		gl.TextureStorage2D(texId, levels, gl.RGBA8, int32(bounds.Max.X),
			int32(bounds.Max.Y))
		err = CheckError()
		if err != nil {
			return 0, err
		}
		gl.TextureSubImage2D(texId, 0, 0, 0, int32(bounds.Max.X),
			int32(bounds.Max.Y), gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
		err = CheckError()
		if err != nil {
			return 0, err
		}
		if levels > 1 {
			gl.BindTexture(gl.TEXTURE_2D, texId)
			var v float32
			gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY, &v)
			gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAX_ANISOTROPY, v)
			gl.GenerateMipmap(gl.TEXTURE_2D)
			gl.BindTexture(gl.TEXTURE_2D, 0)
		}
		return texId, nil
	}
}

func loadTimestamps(doc *gltf.Document, acc *gltf.Accessor, data map[int][]byte) ([]float32, error) {
	accb, err := accessorAttrib(doc, acc, data)
	if err != nil {
		return nil, err
	}
	if accb.size != int(unsafe.Sizeof(float32(0))) {
		return nil, errors.New("invalid timestamp type")
	}
	timestamps := make([]float32, 0, accb.count)
	for i := range accb.count {
		e := accb.src[i*accb.size:]
		val := binary.LittleEndian.Uint32(e)
		timestamps = append(timestamps, math.Float32frombits(val))
	}
	return timestamps, nil
}

func loadInverseBindMats(doc *gltf.Document, acc *gltf.Accessor, data map[int][]byte) ([]y3d.Mat4, error) {
	a, err := accessorAttrib(doc, acc, data)
	if err != nil {
		return nil, err
	}
	if a.size != int(unsafe.Sizeof(y3d.Mat4{})) {
		return nil, errors.New("invalid timestamp type")
	}
	out := make([]y3d.Mat4, a.count)
	ptr := unsafe.Pointer(reflect.ValueOf(a.src).Index(0).UnsafeAddr())
	in := unsafe.Slice((*y3d.Mat4)(ptr), a.count)
	copy(out, in)
	return out, nil
}

// add binding index to this
func attributeNames[V any](attrs map[string]V) ([]string, []uint32) {
	rank := func(name string) int {
		for k, n := range attribOrder {
			if n == name {
				return k
			}
		}
		return len(attribOrder)
	}
	names := make([]string, 0, len(attrs))
	indices := make([]uint32, 0, len(attrs))
	for name := range attrs {
		indices = append(indices, uint32(rank(name)))
		names = append(names, name)
	}
	sort.Slice(names, func(a, b int) bool {
		ra, rb := rank(names[a]), rank(names[b])
		if ra != rb {
			return ra < rb
		}
		return names[a] < names[b]
	})
	slices.Sort(indices)
	return names, indices
}

func align4(n int) int { return (n + 3) &^ 3 }

func makeVBuffer(v, p *bytes.Buffer,
	primIndx int, format []VertexFormat, attribs []vertexAttrib) (int, int, error) {
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
	//copy the primitive index for each vertex
	for range count {
		binary.Write(p, binary.LittleEndian, uint32(primIndx))
	}
	v.Write(out)
	out = nil
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
func readAccessor[T any](doc *gltf.Document, idx int, data map[int][]byte, want gltf.AccessorType) ([]T, error) {
	acc := doc.Accessors[idx]
	if acc.ComponentType != gltf.ComponentFloat || acc.Type != want {
		return nil, fmt.Errorf("accessor %d: want float %v, got %v %v",
			idx, want, acc.ComponentType, acc.Type)
	}
	if acc.Count == 0 {
		return nil, nil
	}
	attrib, err := accessorAttrib(doc, acc, data)
	if err != nil {
		return nil, fmt.Errorf("accessor %d: %w", idx, err)
	}
	addr := unsafe.Pointer(reflect.ValueOf(attrib.src).Index(0).UnsafeAddr())
	out := make([]T, acc.Count)
	copy(out, unsafe.Slice((*T)(addr), acc.Count))
	return out, nil
}

func readTrack[T any](doc *gltf.Document, s *gltf.AnimationSampler, data map[int][]byte, typ gltf.AccessorType) (Track[T], error) {
	times, err := readAccessor[float32](doc, s.Input, data, gltf.AccessorScalar)
	if err != nil {
		return Track[T]{}, err
	}
	values, err := readAccessor[T](doc, s.Output, data, typ)
	if err != nil {
		return Track[T]{}, err
	}
	want := len(times)
	if s.Interpolation == gltf.InterpolationCubicSpline {
		want *= 3
	}
	if len(values) != want {
		return Track[T]{}, fmt.Errorf("sampler has %d keys but %d values (want %d)",
			len(times), len(values), want)
	}
	return Track[T]{Times: times, Values: values}, nil
}

func loadSkeleton(doc *gltf.Document, skin *gltf.Skin, data map[int][]byte) ([]*Joint, map[int]int, error) {
	n := len(skin.Joints)
	mats := make([]y3d.Mat4, n)
	for i := range mats {
		mats[i] = y3d.Identity
	}
	if skin.InverseBindMatrices != nil {
		m, err := loadInverseBindMats(doc, doc.Accessors[*skin.InverseBindMatrices], data)
		if err != nil {
			return nil, nil, err
		}
		if len(m) < n {
			return nil, nil, fmt.Errorf("skin has %d joints but %d inverse bind matrices", n, len(m))
		}
		copy(mats, m)
	}
	jointsMap := make(map[int]int, n)
	joints := make([]*Joint, n)
	for i, j := range skin.Joints {
		joints[i] = NewJoint(j, -1, doc.Nodes[j].Name, mats[i])
		jointsMap[j] = i
	}
	for i, j := range skin.Joints {
		for _, c := range doc.Nodes[j].Children {
			if ci, ok := jointsMap[c]; ok {
				joints[ci].Parent = i
				joints[i].Children = append(joints[i].Children, ci)
			}
		}
	}
	return joints, jointsMap, nil
}

func newSampler(id int, n *gltf.Node) *Sampler {
	t, r, s := n.TranslationOrDefault(), n.RotationOrDefault(), n.ScaleOrDefault()
	return &Sampler{
		Id:             id,
		Repeat:         true,
		OutTranslation: y3d.Vec3{X: float32(t[0]), Y: float32(t[1]), Z: float32(t[2])},
		OutRotation:    y3d.Quaternion{W: float32(r[0]), X: float32(r[1]), Y: float32(r[2]), Z: float32(r[3])}, // x, y, z, w
		OutScale:       y3d.Vec3{X: float32(s[0]), Y: float32(s[1]), Z: float32(s[2])},
	}
}

func loadAnimations(doc *gltf.Document, jointsMap map[int]int, data map[int][]byte) ([]*Animation, error) {
	anims := make([]*Animation, 0, len(doc.Animations))
	for ai, a := range doc.Animations {
		if a == nil {
			continue
		}
		anim := NewAnimation()
		for ci, c := range a.Channels {
			if c.Target.Node == nil {
				continue
			}
			ji, ok := jointsMap[*c.Target.Node]
			if !ok {
				continue // targets a node outside this skin
			}
			s := anim.Samplers[ji]
			if s == nil {
				s = newSampler(ji, doc.Nodes[*c.Target.Node])
				anim.Samplers[ji] = s
			}

			gs := a.Samplers[c.Sampler]
			var err error
			switch c.Target.Path {
			case gltf.TRSRotation:
				s.rotation, err = readTrack[y3d.Quaternion](doc, gs, data, gltf.AccessorVec4)
				s.Flag |= SAMPLER_ROTATION
			case gltf.TRSTranslation:
				s.translation, err = readTrack[y3d.Vec3](doc, gs, data, gltf.AccessorVec3)
				s.Flag |= SAMPLER_TRANSLATION
			case gltf.TRSScale:
				s.scale, err = readTrack[y3d.Vec3](doc, gs, data, gltf.AccessorVec3)
				s.Flag |= SAMPLER_SCALE
			default:
				continue // morph target weights not supported
			}
			if err != nil {
				return nil, fmt.Errorf("animation %d (%q) channel %d: %w", ai, a.Name, ci, err)
			}
		}
		anims = append(anims, anim)
	}
	return anims, nil
}

func ProcessNode(world *ywd.World, node *gltf.Node,
	doc *gltf.Document,
	r *RenderManager,
	skin int,
	data map[int][]byte) error {
	if node.Mesh != nil {
		ynode := world.NewEntity()
		mesh := doc.Meshes[*node.Mesh]
		v, i, p := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
		vertexType := ""
		var format []VertexFormat
		var base uint32
		var vertexSize int
		pbrMaterialData := make([]PbrMaterialData, 0)
		pbrMaterials := make([]PbrMaterial, 0)
		for primIndx, primitive := range mesh.Primitives {
			names, indices := attributeNames(primitive.Attributes)
			pf := make([]VertexFormat, 0, len(names))
			attribs := make([]vertexAttrib, 0, len(names))
			for i, name := range names {
				acc := doc.Accessors[primitive.Attributes[name]]
				a, err := accessorAttrib(doc, acc, data)
				if err != nil {
					return fmt.Errorf("attribute %s: %w", name, err)
				}
				pf = append(pf, VertexFormat{
					ComponentSize: int32(acc.Type.Components()),
					Type:          uint32(compMap[acc.ComponentType]),
					Normalized:    acc.Normalized,
					BindingIndex:  indices[i],
				})
				attribs = append(attribs, a)
			}
			vertexType = strings.Join(names, ".")
			count, vs, err := makeVBuffer(v, p, primIndx, pf, attribs)
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
			if primitive.Material != nil {
				material := doc.Materials[*primitive.Material]
				materialData := PbrMaterialData{
					BaseColorFactor: y3d.Vec4{
						X: float32(material.PBRMetallicRoughness.BaseColorFactor[0]),
						Y: float32(material.PBRMetallicRoughness.BaseColorFactor[1]),
						Z: float32(material.PBRMetallicRoughness.BaseColorFactor[2]),
						W: float32(material.PBRMetallicRoughness.BaseColorFactor[3]),
					},
					EmissiveFactor: y3d.Vec3{
						X: float32(material.EmissiveFactor[0]),
						Y: float32(material.EmissiveFactor[1]),
						Z: float32(material.EmissiveFactor[2]),
					},
				}
				if material.PBRMetallicRoughness.MetallicFactor != nil {
					materialData.MetallicFactor = float32(*material.PBRMetallicRoughness.MetallicFactor)
				}
				if material.PBRMetallicRoughness.RoughnessFactor != nil {
					materialData.RoughnessFactor = float32(*material.PBRMetallicRoughness.RoughnessFactor)
				}
				var pbrMaterial PbrMaterial
				if material.PBRMetallicRoughness.BaseColorTexture != nil {
					texIndx := material.PBRMetallicRoughness.BaseColorTexture.Index
					texId, err := loadTexture(doc, texIndx, r.TextureManager)
					if err != nil {
						return err
					}
					pbrMaterial.Albeo = texId
				}
				if material.PBRMetallicRoughness.MetallicRoughnessTexture != nil {
					texIndx := material.PBRMetallicRoughness.MetallicRoughnessTexture.Index
					texId, err := loadTexture(doc, texIndx, r.TextureManager)
					if err != nil {
						return err
					}
					pbrMaterial.Specular = texId
				}
				if material.NormalTexture != nil {
					texIndx := material.NormalTexture.Index
					if texIndx != nil {
						texId, err := loadTexture(doc, *texIndx, r.TextureManager)
						if err != nil {
							return err
						}
						pbrMaterial.NormalMap = texId
					}
				}

				pbrMaterialData = append(pbrMaterialData, materialData)
				pbrMaterials = append(pbrMaterials, pbrMaterial)
			}
			base += uint32(count)
		}
		tempVertexType := GetVertexFormat(format, r)
		if tempVertexType == INVALID_VERTEX_FORMAT {
			//EACH PRIMITVE CAN HAVE DIFFERENT DRAW MODES
			r.VertextManager.CreateVertexCache(vertexType, gl.TRIANGLES, vertexSize, format)
		} else {
			vertexType = tempVertexType
		}

		position := y3d.Vec3{
			X: float32(node.Translation[0]),
			Y: float32(node.Translation[1]),
			Z: float32(node.Translation[2]),
		}
		rotation := y3d.Quaternion{
			X: float32(node.Rotation[0]),
			Y: float32(node.Rotation[1]),
			Z: float32(node.Rotation[2]),
			W: float32(node.Rotation[3]),
		}
		err := AddEntityRenderableData(
			world,
			ynode,
			vertexType,
			1.0,
			v, i, p,
			r.VertextManager,
		)
		if err != nil {
			return err
		}
		AddEntityMoveData(
			world,
			ynode,
			position,
			rotation,
		)
		world.AddComponent(ynode, ywd.PbrMaterialComponent, pbrMaterials)
		world.AddComponent(ynode, ywd.PbrMaterialDataComponent, pbrMaterialData)
		animsBySkin := map[int][]*Animation{}
		if node.Skin != nil {
			joints, jointsMap, err := loadSkeleton(doc, doc.Skins[*node.Skin], data)
			if err != nil {
				return err
			}
			world.AddComponent(ynode, ywd.SkeletonComponent, NewSkeleton(int(ynode), joints))
			anims, ok := animsBySkin[*node.Skin]
			if !ok {
				if anims, err = loadAnimations(doc, jointsMap, data); err != nil {
					return err
				}
				animsBySkin[*node.Skin] = anims
			}
			world.AddComponent(ynode, ywd.AnimationComponent, anims)
		}

	}

	for _, n := range node.Children {
		err := ProcessNode(world, doc.Nodes[n], doc, r, skin, data)
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
		if sameFormat(vf, f) {
			return vt
		}
	}
	return INVALID_VERTEX_FORMAT
}
