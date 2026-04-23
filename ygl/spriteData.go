package ygl

import (
	"unsafe"
	"yam/y3d"
)

var (
	SpriteData = [20]float32{
		1.0, 1.0, 1.0, 1.0, 1.0, // top-right    [0]
		1.0, -1.0, 1.0, 1.0, 0.0, // bottom-right [1]
		-1.0, -1.0, 1.0, 0.0, 0.0, // bottom-left  [2]
		-1.0, 1.0, 1.0, 0.0, 1.0, // top-left     [3]
	}
	SpriteFormat = [2]DataFormat{
		{
			Count:         3,
			Stride:        int32(uintptr(5) * unsafe.Sizeof(float32(0))),
			Offset:        0,
			ComponentType: ComponentTypeFloat32,
		},
		{
			Count:         2,
			Stride:        int32(uintptr(5) * unsafe.Sizeof(float32(0))),
			Offset:        int32(uintptr(3) * unsafe.Sizeof(float32(0))),
			ComponentType: ComponentTypeFloat32,
		},
	}
	SpriteIndices []uint16 = []uint16{
		0, 1, 3, 1, 2, 3,
	}

	SpriteVert string = `#version 330
	 layout(location = 0) in vec3 pos;
	 layout(location = 1) in vec2 uv;
	 
	 uniform mat4 projView;
	 uniform mat4 world;
	 
	 out vec2 frag_uv;
	
	 
	 void main() {
	 	frag_uv = uv;
		gl_Position =  projView * world * vec4(pos, 1);								  
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

func MakeAABBForSprite(SpriteData []float32, vertFormat DataFormat) y3d.AABB {
	stride := int(uintptr(vertFormat.Stride) / unsafe.Sizeof(float32(0)))
	firstPoint := y3d.Vec3{
		X: float32(SpriteData[0]),
		Y: float32(SpriteData[1]),
		Z: float32(SpriteData[2]),
	}
	aabb := y3d.AABB{
		Min: firstPoint,
		Max: firstPoint,
	}
	for i := range 4 {
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
