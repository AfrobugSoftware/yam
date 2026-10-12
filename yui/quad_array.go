package yui

import (
	"errors"
	"fmt"
	"unsafe"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	MAX_QUADS      = 10000
	BUFFER_BINDING = 15
)

var (
	vertShader = `#version 430 core

#define MAX_QUADS 10000
const vec2 positions[6] = vec2[6](
	    vec2(0.0, 0.0),
		vec2(0.0, 1.0),
		vec2(1.0, 1.0),
		vec2(0.0, 0.0),
		vec2(1.0, 1.0),
		vec2(1.0, 0.0));


layout (std430, binding = 15) buffer QuadInfo {
    vec4 BasePosWidthHeight[MAX_QUADS];
    vec4 TexCoordsWidthHeight[MAX_QUADS];
    vec4 Color[MAX_QUADS];
	uint Flag[MAX_QUADS];
};

out Frag {
    vec4 Color;
    vec2 TexCoords;
	flat uint Flag;
	flat float Idx;
}vs_out;

layout(location = 1) uniform mat4 proj;

void main() {
    vec3 basePos = vec3(BasePosWidthHeight[gl_InstanceID].xy, 0.5);
    vec2 widthHeight = positions[gl_VertexID] * BasePosWidthHeight[gl_InstanceID].zw;
    vec3 newPosition = basePos +  vec3(widthHeight, 0.0);
    gl_Position = proj * vec4(newPosition, 1.0);

    vec2 baseTexCoords = TexCoordsWidthHeight[gl_InstanceID].xy;
    vec2 texWidthHeight = positions[gl_VertexID] * TexCoordsWidthHeight[gl_InstanceID].zw;
    vec2 aTexCoord = baseTexCoords + texWidthHeight;
    
    vs_out.TexCoords = vec2(aTexCoord.x, aTexCoord.y);
    vs_out.Color = Color[gl_InstanceID];

	vs_out.Flag = Flag[gl_InstanceID];
	vs_out.Idx = Color[gl_InstanceID].a;
}`
	fragShader = `#version 430 core 

layout (location = 0) out vec4 FragColor;

uniform sampler2D tex;
layout (binding = 1) uniform sampler2DArray fontTex;
in Frag {
    vec4 Color;
    vec2 TexCoords;
	flat uint Flag;
	flat float Idx;
}fs_in;

void main () {
    // FragColor = texture2D(tex, fs_in.TexCoords.xy);
    // if (FragColor == vec4(0.0)) {
	// }
	if((fs_in.Flag & (1 << 18)) != 0) {
		FragColor = vec4(fs_in.Color.rgb, texture(fontTex, vec3(fs_in.TexCoords.xy, fs_in.Idx)).r);
	}else{
		FragColor = fs_in.Color;
	}
}
`
)

type QuadArray struct {
	vbo                        uint32
	ssbo                       uint32
	blockSize                  int32
	basePosWidthHeightOffset   int32
	texCoordsWidthHeightOffset int32
	colorOffset                int32
	flagOffset                 int32
	blockPtr                   unsafe.Pointer
}

func CreateQuadArray(shader uint32) *QuadArray {
	var vbo uint32
	gl.CreateVertexArrays(1, &vbo)
	gl.BindVertexArray(vbo)
	q := &QuadArray{
		vbo: vbo,
	}
	q.basePosWidthHeightOffset = 0
	q.texCoordsWidthHeightOffset = int32(unsafe.Sizeof(y3d.Vec4{})) * MAX_QUADS
	q.colorOffset = (int32(unsafe.Sizeof(y3d.Vec4{})) * MAX_QUADS) * 2
	q.flagOffset = (int32(unsafe.Sizeof(y3d.Vec4{})) * MAX_QUADS) * 3
	q.blockSize = (int32(unsafe.Sizeof(y3d.Vec4{})*MAX_QUADS))*3 + (int32(unsafe.Sizeof(uint32(0))) * MAX_QUADS)
	gl.CreateBuffers(1, &q.ssbo)
	gl.NamedBufferStorage(q.ssbo, int(q.blockSize), nil, gl.DYNAMIC_STORAGE_BIT|gl.MAP_WRITE_BIT)
	return q
}

func (q *QuadArray) Draw(numQuads int, font *Font) {
	if numQuads <= 0 {
		return
	} else if numQuads >= MAX_QUADS {
		numQuads = MAX_QUADS - 1
	}
	if font != nil {
		gl.BindTextureUnit(1, font.TextureID)
	}
	gl.BindBufferBase(gl.SHADER_STORAGE_BUFFER, BUFFER_BINDING, q.ssbo)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
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
	texUSize, texVSize, r, g, b, a float32, flag uint32) {
	if idx >= MAX_QUADS {
		return
	}
	if q.blockPtr == nil {
		return
	}

	block := unsafe.Slice((*byte)(q.blockPtr), q.blockSize)
	basePos := ViewAsFloat32(block, int(q.basePosWidthHeightOffset), MAX_QUADS)
	texCoords := ViewAsFloat32(block, int(q.texCoordsWidthHeightOffset), MAX_QUADS)
	color := ViewAsFloat32(block, int(q.colorOffset), MAX_QUADS)
	flagb := ViewAsUint32(block, int(q.flagOffset), MAX_QUADS)

	basePos[idx][0] = x
	basePos[idx][1] = y
	basePos[idx][2] = width
	basePos[idx][3] = height
	texCoords[idx][0] = UBase
	texCoords[idx][1] = VBase
	texCoords[idx][2] = texUSize
	texCoords[idx][3] = texVSize
	color[idx][0] = r
	color[idx][1] = g
	color[idx][2] = b
	color[idx][3] = a
	flagb[idx] = flag
}

func ViewAsFloat32(buf []byte, byteOffset, count int) [][4]float32 {
	const pairSize = int(unsafe.Sizeof([4]float32{}))
	required := byteOffset + count*pairSize
	if required > len(buf) {
		panic(fmt.Errorf("ViewAsFloat32Pairs: slice out of bounds, req: %d, len : %d\n", required, len(buf)))
	}
	ptr := unsafe.Pointer(&buf[byteOffset])
	return unsafe.Slice((*[4]float32)(ptr), count)
}
func ViewAsUint32(buf []byte, byteOffset, count int) []uint32 {
	const pairSize = int(unsafe.Sizeof(uint32(0)))
	required := byteOffset + count*pairSize
	if required > len(buf) {
		panic(fmt.Errorf("ViewAsFloat32Pairs: slice out of bounds, req: %d, len : %d\n", required, len(buf)))
	}
	ptr := unsafe.Pointer(&buf[byteOffset])
	return unsafe.Slice((*uint32)(ptr), count)
}
