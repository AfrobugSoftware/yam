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
	gbufferVertex = `
	#version 430 core

layout(location = 0) in vec3 pos;
layout(location = 1) in vec3 normal;
layout(location = 2) in vec2 uv;
layout(location = 3) in vec4 color;
layout(location = 10) in uint dInx;
out data {
	vec4 fragPos;
	vec3 fragNormal;
	vec4 fragColor;
	vec2 fragUv;

}vs_out;
layout(std430, binding = 0) buffer MatrixSSBO {
    	mat4 world[];
	};
	layout(std140, binding = 4) uniform vp {
		mat4 viewproj;
	}

void main() {
    vs_in.fragUv = uv;
    vs_in.fragPos = world[dInx] * vec4(pos, 1.0);
    vs_in.fragNormal = mat3(transpose(inverse(world[dInx]))) * normal; 
    vs_in.fragColor = color;
	gl_Position = viewproj * world * vec4(pos, 1.0);
}
	`

	gbufferFragment = `
	#version 430 core

layout (location = 0) out vec4 outWolrdPos;
layout (location = 1) out vec4 outWorldNormal;
layout (location = 2) out vec4 outDiffuse;

in data {
    vec4 fragPos;
    vec3 fragNormal;
	vec4 fragColor
    vec2 fragUv;
}fs_in;


layout(binding = 0) uniform sampler2D mdiffuse;

void main() {
	vec4 out_color = texture(mdiffuse, fs_in.fragUv.xy);
	if(out_color == vec4(0.0)) {
		outDiffuse = fs_in.fragColor;
	}else{
		outDiffuse = out_color
	}
    outDiffuse = out_color;
    outWolrdPos = fs_in.fragPos;
    outWorldNormal = vec4(fs_in.fragNormal, 0.0);
}
	`
	screenquadVertex = `
	#version 430 core

const vec3 vertices[6] = vec3[6](
    vec3(-1.0, -1.0, 0.5),
    vec3(1.0, -1.0, 0.5),
    vec3(1.0,1.0,0.5),

    vec3(-1.0, -1.0,0.5),
    vec3(1.0,1.0,0.5),
    vec3(-1.0, 1.0, 0.5)
);

out vec2 fragUv;

void main() {
    vec4 pos = vec4(vertices[gl_VertexID], 1.0);
    gl_Position = pos;
    fragUv = vec2((pos.x +1) / 2.0, (pos.y + 1) / 2.0);
}
	`
)
