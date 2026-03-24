package ygl

import (
	"unsafe"

	gl "github.com/chsc/gogl/gl33"
)

var (
	// SpriteData [20]gl.Float = [20]gl.Float{
	// 	1.0, 1.0, 0.0, 1.0, 1.0,
	// 	1.0, -1.0, 0.0, 0.0, 1.0,
	// 	-1.0, -1.0, 0.0, 0.0, 0.0,
	// 	-1.0, 1.0, 0.0, 0.0, 1.0,
	// }
	SpriteData [20]gl.Float = [20]gl.Float{
		0.5, 0.5, 0.0, 1.0, 1.0, // top-right    [0]
		0.5, -0.5, 0.0, 1.0, 0.0, // bottom-right [1]
		-0.5, -0.5, 0.0, 0.0, 0.0, // bottom-left  [2]
		-0.5, 0.5, 0.0, 0.0, 1.0, // top-left     [3]
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
		//frag_color = texture(tex, frag_uv);
		frag_color = vec4(1.0, 1.0,0.0,1.0);
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
