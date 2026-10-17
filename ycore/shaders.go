package ycore

const (
	blingPhongVertex = `
#version 430
layout(location = 0) in vec3 pos;
layout(location = 1) in vec3 norm;
layout(location = 2) in vec2 tc;
layout(location = 3) in vec4 color;
layout(location = 10) in uint dInx;

struct per_object_transform {
	mat4 world;
	mat4 normal;
};

layout(std430, binding = 0) buffer matrix_ssbo {
    per_object_transform world[];
};
layout(std140, binding = 4) uniform vp {
	mat4 view_proj;
};
out data {
	vec3 pos;
	vec3 normal;
	vec2 tc;
	vec4 color;
}vs_out;
void main() {
	vs_out.pos = vec3(world[dInx].world * vec4(pos, 1.0));
	vs_out.normal = mat3(world[dInx].normal) * norm;
	vs_out.tc = tc;
	gl_Position = view_proj * world[dInx].world * vec4(pos, 1.0);
};
	`
	blingPhongFragment = `
#version 430
layout(location = 0) out vec4 out_color;
in data {
	vec3 pos;
	vec3 normal;
	vec2 tc;
	vec4 color;
}fs_in;

struct light {
	vec4 pos;
	vec4 color;
	float radius;
	float inner_radius;
	uint flag;
};

layout(std140, binding= 2) uniform material_s {
    vec4 diffuse;
    vec4 ambient;
    vec4 specular;
    vec4 emissive_shineness;
} material;

layout(std430,binding = 3) buffer render_parameters {
	vec3 view_pos;
	float game_time;
	light l[128];
}render_params;

vec4 calculate_bling_phong(int light_index) {
	vec3 light_direction = normalize(vec3(render_params.l[light_index].pos));
	vec3 view_direction = normalize(render_params.view_pos - fs_in.pos);
	vec3 halfway_direction = normalize(light_direction + view_direction);

	float specular_factor = pow(max(dot(fs_in.normal, halfway_direction), 0.0), material.emissive_shineness.w);
	vec4 specular = render_params.l[light_index].color * material.specular * specular_factor;

	float diffuse_factor = max(dot(fs_in.normal, light_direction), 0.0);
	vec4 diffuse = render_params.l[light_index].color * material.diffuse * diffuse_factor;

	return (diffuse + specular);
}

void main() {
	out_color = vec4(calculate_bling_phong(0).xyz, 1.0);
}	
	`
	shadowVertexShader = `
#version 430
layout(location = 0) in vec3 pos;
layout(location = 10) in uint dInx;
layout(std430, binding = 0) buffer MatrixSSBO {
    mat4 world[];
};
layout(std140, binding = 4) uniform vp {
	mat4 viewproj;
};
void main() {
	gl_Position = viewproj * world[dInx] * vec4(pos, 1.0);
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
layout(std430, binding = 0) buffer matrix {
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
