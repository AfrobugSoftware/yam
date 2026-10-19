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
	flat uint draw_index;
}vs_out;
void main() {
	vs_out.pos = vec3(world[dInx].world * vec4(pos, 1.0));
	vs_out.normal = mat3(world[dInx].normal) * norm;
	vs_out.tc = tc;
	vs_out.draw_index = dInx;
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
	flat uint draw_index;
}fs_in;

struct light {
	vec4 pos;
	vec4 color;
	float radius;
	float inner_radius;
	uint flag;
};

struct per_object_render_state {
	uint render_flag;
	float alpha;
};

layout(std430, binding = 2) buffer per_object_render_state_buffer {
	per_object_render_state data[];
}ors;

layout(std140, binding= 2) uniform material_s {
    vec4 diffuse;
    vec4 ambient;
    vec4 specular;
    vec4 emissive_shineness;
} material;

layout(std430,binding = 3) buffer render_parameters {
	vec3 view_pos;
	float game_time;
	mat4 shadow_matrix;
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
	vec4 ambient = material.ambient * render_params.l[light_index].color * 0.1;
	return (diffuse + specular + ambient);
}

void main() {
	out_color = vec4(calculate_bling_phong(0).xyz, ors.data[fs_in.draw_index].alpha);
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
	vec4 pos;
	vec3 normal;
	vec2 tc;
	vec4 color;
	flat uint draw_index;
}vs_out;

void main() {
    vs_out.tc     = tc;
    vs_out.pos    = world[dInx].world * vec4(pos, 1.0);
    vs_out.normal = mat3(world[dInx].normal) * norm; 
    vs_out.color  = color;
	gl_Position = viewproj * world[dInx].world * vec4(pos, 1.0);
}
	`

	gbufferFragment = `
#version 430 core

layout (location = 0) out vec4 positions;
layout (location = 1) out vec4 normals;
layout (location = 2) out vec4 albedo;
in data {
	vec4 pos;
	vec3 normal;
	vec2 tc;
	vec4 color;
	flat uint draw_index;
}fs_in;

layout(binding = 0) uniform sampler2D mdiffuse;

void main() {
    albedo = fs_in.color;
    positions = fs_in.pos;
    normals = vec4(fs_in.normal, 0.0);
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
