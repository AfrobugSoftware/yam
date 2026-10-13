package ycore

const (
	shadowVertexShader = `
	#version 430
	layout(location = 0) in vec3 pos;
	layout(location = 10) in uint dInx;
	layout(std430, binding = 0) buffer MatrixSSBO {
    	mat4 world[];
	};
	layout(std140, binding = 4) uniform vp {
		mat4 viewproj;
	}
	void main() {
		gl_Position = viewproj * world[dInx] * pos;
	};
	`
	shadowFragmentShader = `
	#version 430

	void main() {
		//
	}
	`
)
