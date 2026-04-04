package ygl

import (
	"unsafe"
	"yam/y3d"

	gl "github.com/chsc/gogl/gl33"
)

var (
	SpriteData [20]gl.Float = [20]gl.Float{
		1.0, 1.0, 1.0, 1.0, 1.0, // top-right    [0]
		1.0, -1.0, 1.0, 1.0, 0.0, // bottom-right [1]
		-1.0, -1.0, 1.0, 0.0, 0.0, // bottom-left  [2]
		-1.0, 1.0, 1.0, 0.0, 1.0, // top-left     [3]
	}
	SpriteFormat [2]DataFormat = [2]DataFormat{
		{
			Count:  3,
			Stride: int(uintptr(5) * unsafe.Sizeof(gl.Float(0))),
			Length: 4,
			Offset: 0,
		},
		{
			Count:  2,
			Stride: int(uintptr(5) * unsafe.Sizeof(gl.Float(0))),
			Length: 4,
			Offset: 3,
		},
	}
	SpriteIndices []uint16 = []uint16{
		0, 1, 3, 1, 2, 3,
	}

	SpriteVert string = `#version 330
	 layout(location = 0) in vec3 pos;
	 layout(location = 1) in vec2 uv;
	 
	 uniform mat4 proj;
	 uniform mat4 world;
	 uniform mat4 view;
	 
	 out vec2 frag_uv;
	
	 
	 void main() {
	 	frag_uv = uv;
		gl_Position =  proj * view * world * vec4(pos, 1);								  
	 }`
	SpriteFrag string = `#version 330
	in vec2 frag_uv;
	out vec4 frag_color;
	
	uniform sampler2D tex;
	
	void main() {
		frag_color = texture(tex, frag_uv);
		//frag_color = vec4(1.0, 1.0,0.0,1.0);
	}`

	SpriteAnimFrag string = `#version 330
	in vec2 frag_uv;
	out vec4 frag_color;
	
	uniform sampler2D tex;
	uniform float frame;
	uniform int frameW;
	unifrom int frameH;

	void main() {
		
	}
	`
)

func MakeAABBForSprite(SpriteData []gl.Float, vertFormat DataFormat) y3d.AABB {
	stride := int(uintptr(vertFormat.Stride) / unsafe.Sizeof(gl.Float(0)))
	firstPoint := y3d.Vec3{
		X: float32(SpriteData[0]),
		Y: float32(SpriteData[1]),
		Z: float32(SpriteData[2]),
	}
	aabb := y3d.AABB{
		Min: firstPoint,
		Max: firstPoint,
	}
	for i := range vertFormat.Length {
		point := y3d.Vec3{
			X: float32(SpriteData[i*stride]),
			Y: float32(SpriteData[i*stride+1]),
			Z: float32(SpriteData[i*stride+2]),
		}
		aabb.Min.X = min(aabb.Min.X, point.X)
		aabb.Min.Y = min(aabb.Min.Y, point.Y)
		aabb.Min.Z = min(aabb.Min.Z, point.Z)

		aabb.Max.X = max(aabb.Max.X, point.X)
		aabb.Max.Y = max(aabb.Max.Y, point.Y)
		aabb.Max.Z = max(aabb.Max.Z, point.Z)
	}
	return aabb
}
