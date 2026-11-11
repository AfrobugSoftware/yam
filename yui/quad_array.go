package yui

import (
	"errors"
	"unsafe"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	BUFFER_BINDING = 15
)

var (
	vertShader = `#version 430 core
const vec2 positions[6] = vec2[6](
	    vec2(0.0, 0.0),
		vec2(0.0, 1.0),
		vec2(1.0, 1.0),
		vec2(0.0, 0.0),
		vec2(1.0, 1.0),
		vec2(1.0, 0.0));

struct quad_info {
	vec4 pos_width_height;
	vec4 texcoord_width_height;
	vec4 color;
	float radius;
	uint flag;
	uint tex;
};

layout (std430, binding = 15) buffer QuadInfo {
   quad_info info[];
};

out Frag {
    vec4 Color;
	vec2 Pos;
    vec2 TexCoords;
    flat vec2 Extents;
	flat uint Flag;
	flat uint Tex;
	flat float Idx;
	flat float Radius;
}vs_out;

layout(location = 1) uniform mat4 proj;

void main() {
    vec3 basePos = vec3(info[gl_InstanceID].pos_width_height.xy, 0.5);
    vec2 widthHeight = positions[gl_VertexID] * info[gl_InstanceID].pos_width_height.zw;
    vec3 newPosition = basePos +  vec3(widthHeight, 0.0);
    gl_Position = proj * vec4(newPosition, 1.0);

    vec2 baseTexCoords = info[gl_InstanceID].texcoord_width_height.xy;
    vec2 texWidthHeight = positions[gl_VertexID] * info[gl_InstanceID].texcoord_width_height.zw;
    vec2 aTexCoord = baseTexCoords + texWidthHeight;
    
    vs_out.TexCoords = vec2(aTexCoord.x, aTexCoord.y);
    vs_out.Color = info[gl_InstanceID].color;
	vs_out.Flag = info[gl_InstanceID].flag;
	vs_out.Idx = info[gl_InstanceID].color.a;
	vs_out.Radius = info[gl_InstanceID].radius;
	vs_out.Pos = positions[gl_VertexID];
	vs_out.Extents = info[gl_InstanceID].pos_width_height.zw;
	vs_out.Tex = info[gl_InstanceID].tex;
}`
	fragShader = `#version 430 core 

layout (location = 0) out vec4 FragColor;

//uniform sampler2D tex;
uniform sampler2D fontTex[32];
in Frag {
    vec4 Color;
	vec2 Pos;
    vec2 TexCoords;
	flat vec2 Extents;
	flat uint Flag;
	flat uint Tex;
	flat float Idx;
	flat float Radius;
}fs_in;


bool InConner(vec2 point, float radius, vec2 dim, out vec2 cen) {
	radius = min(min(radius, dim.x/2),dim.y);
	if(point.x < radius) {
		cen.x = radius;
	}else if (point.x > dim.x-radius) {
		cen.x = dim.x-radius;
	}else{
		return false;
	}
	if(point.y < radius){
		cen.y = radius;
	}else if (point.y > dim.y - radius){
		cen.y = dim.y - radius;
	}else{
		return false;
	}
	return true;
}

void main () {
	if((fs_in.Flag & (1 << 18)) != 0) {
		FragColor = vec4(fs_in.Color.rgb, texture(fontTex[fs_in.Tex], fs_in.TexCoords.xy).r);
	}else{
		float u_radius = fs_in.Radius;
		vec2 u_dimensions = fs_in.Extents;
		vec2 coords = fs_in.Pos * (fs_in.Extents);
		vec2 cen;
		if (InConner(coords, u_radius, u_dimensions, cen)) {
			if(length(coords - cen) <= u_radius) {
				FragColor = fs_in.Color;	
			}else {
				discard;
			}
		}else{
			if((fs_in.Flag & (1 << 20)) != 0) {
				FragColor = texture(fontTex[fs_in.Tex], fs_in.TexCoords.xy);
			}else{
				FragColor = fs_in.Color;	
			}
		}
	}
}
`
)

type QuadInfo struct {
	PosWidthHeight      y3d.Vec4
	TexCoordWidthHeight y3d.Vec4
	Color               y3d.Vec4
	Radius              float32
	Flag                uint32
	Tex                 uint32
	padding             uint32
}

type QuadArray struct {
	vbo       uint32
	ssbo      uint32
	blockSize int32
	blockPtr  unsafe.Pointer
	texs      []uint32
}

func CreateQuadArray(shader uint32) *QuadArray {
	var vbo uint32
	gl.CreateVertexArrays(1, &vbo)
	gl.BindVertexArray(vbo)
	var uints int32
	gl.GetIntegerv(gl.MAX_TEXTURE_IMAGE_UNITS, &uints)
	q := &QuadArray{
		vbo:  vbo,
		texs: make([]uint32, 0, uints),
	}
	q.blockSize = 1000 * int32(unsafe.Sizeof(QuadInfo{}))
	gl.CreateBuffers(1, &q.ssbo)
	gl.NamedBufferStorage(q.ssbo, int(q.blockSize), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	return q
}

func (q *QuadArray) AddTexture(texId uint32) (int, error) {
	if texId == 0 {
		return 0, errors.New("cannot add a 0 texture")
	}
	var uints int32
	gl.GetIntegerv(gl.MAX_TEXTURE_IMAGE_UNITS, &uints)
	i := len(q.texs)
	if i >= int(uints) {
		return 0, errors.New("no space for additional texture")
	}
	q.texs = append(q.texs, texId)
	return i, nil
}

func (q *QuadArray) Draw(numQuads int) {
	count := q.blockSize / int32(unsafe.Sizeof(QuadInfo{}))
	if numQuads <= 0 {
		return
	} else if numQuads >= int(count) {
		numQuads = int(count) - 1
	}
	for i, t := range q.texs {
		gl.BindTextureUnit(uint32(i), t)
	}

	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, BUFFER_BINDING, q.ssbo)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.Enable(gl.CULL_FACE)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.BindVertexArray(q.vbo)
	gl.DrawArraysInstanced(gl.TRIANGLES, 0, 6, int32(numQuads))
	gl.BindVertexArray(0)

	gl.Disable(gl.BLEND)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
}

func DestroyQuadArray(q *QuadArray) {
	gl.DeleteVertexArrays(1, &q.vbo)
	gl.DeleteBuffers(1, &q.ssbo)
}

func (q *QuadArray) OpenArray() error {
	q.blockPtr = gl.MapNamedBufferRange(
		q.ssbo,
		0,
		int(q.blockSize),
		gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_RANGE_BIT)
	if q.blockPtr == nil {
		return errors.New("cannot map ssbo")
	}
	return nil
}

func (q *QuadArray) CloseArray() {
	gl.UnmapNamedBuffer(q.ssbo)
	q.blockPtr = nil
}

func (q *QuadArray) SetQuad(idx int, x, y,
	width, height,
	UBase, VBase,
	texUSize, texVSize, r, g, b, a, radius float32, flag, tex uint32) {
	count := q.blockSize / int32(unsafe.Sizeof(QuadInfo{}))
	if idx >= int(count) {
		q.CloseArray()
		blockSize := q.blockSize * 2
		var s uint32
		gl.CreateBuffers(1, &s)
		gl.NamedBufferStorage(s, int(blockSize), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
		gl.NamedCopyBufferSubDataEXT(q.ssbo, s, 0, 0, int(blockSize))
		gl.DeleteBuffers(1, &q.ssbo)
		q.ssbo = s
		q.blockSize = blockSize
		count = q.blockSize / int32(unsafe.Sizeof(QuadInfo{}))
		q.OpenArray()
	}
	if q.blockPtr == nil {
		return
	}

	quadinfo := QuadInfo{
		PosWidthHeight: y3d.Vec4{
			X: x,
			Y: y,
			Z: width,
			W: height,
		},
		TexCoordWidthHeight: y3d.Vec4{
			X: UBase,
			Y: VBase,
			Z: texUSize,
			W: texVSize,
		},
		Color: y3d.Vec4{
			X: r,
			Y: g,
			Z: b,
			W: a,
		},
		Radius: radius,
		Flag:   flag,
		Tex:    tex,
	}
	block := unsafe.Slice((*QuadInfo)(q.blockPtr), count)
	block[idx] = quadinfo
}
