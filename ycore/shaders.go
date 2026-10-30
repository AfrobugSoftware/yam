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
	float alpha;
	uint render_flag;
	uint tex1;
	uint tex2;
	uint tex3;
	uint tex4;
	uint base_joint_idx;
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
	gridVertexShader = `
#version 430
const vec3 positions[4] = vec3[4](
	vec3(-1.0, 0.0, -1.0),
	vec3(1.0, 0.0, -1.0),
	vec3(1.0, 0.0,  1.0),
	vec3(-1.0, 0.0, 1.0));
	
const int indices[6] = int[6](0, 2, 1,  0, 3, 2);
layout(std140, binding = 4) uniform vp {
	mat4 viewproj;
};
struct light {
	vec4 pos;
	vec4 color;
	float radius;
	float inner_radius;
	uint flag;
};
layout(std430,binding = 3) buffer render_parameters {
	vec3 view_pos;
	float game_time;
	mat4 shadow_matrix;
	light l[128];
}render_params;

out data{
	vec3 pos;
	flat vec2 campos;
}vs_out;
uniform float grid_size = 100.0;
void main() {
	int index = indices[gl_VertexID];
	vec3 pos = positions[index] * grid_size;
	pos.x += render_params.view_pos.x;
	pos.z += render_params.view_pos.z;
	vs_out.campos = render_params.view_pos.xz;
	vs_out.pos = pos;
	gl_Position = viewproj * vec4(pos, 1.0);
}	
`
	gridFragmentShader = `
#version 430
layout(location = 0) out vec4 out_color;
in data {
	vec3 pos;
	flat vec2 campos;
}fs_in;
uniform float min_pixels_between_cells = 2.0;
uniform float cell_size = 0.025;
uniform vec4  color_thin = vec4(0.5,0.5,0.5,1.0);
uniform vec4  color_thick = vec4(0.0, 0.0,0.0,1.0);

float log10(float x) {
	return log(x)/log(10);
}

vec2 calcAlpha(vec2 position, float grid_size_lod, vec2 dudv) {
	vec2 a = mod(position, grid_size_lod)/ (dudv);
	vec2 b = clamp(a, 0.0, 1.0);
	vec2 c = abs(b * 2 - vec2(1.0));
	vec2 d = vec2(1.0) - c;
	return d;
}
uniform float grid_size = 100.0;
void main() {
	vec2 dvx = vec2(dFdx(fs_in.pos.x), dFdy(fs_in.pos.x));
	vec2 dvy = vec2(dFdx(fs_in.pos.z), dFdy(fs_in.pos.z));
	float lx = length(dvx);
	float ly = length(dvy);
	vec2 dudv = vec2(lx, ly);
	float l = length(dudv);
	float LOD = max(0.0, log10(l * min_pixels_between_cells / cell_size) + 1);
	float grid_size_lod0 = cell_size * pow(10, floor(LOD));
	float grid_size_lod1 = grid_size_lod0 * 10.0;
	float grid_size_lod2 = grid_size_lod1 * 10.0;
	dudv *= 4.0;

	vec2 d = calcAlpha(fs_in.pos.xz, grid_size_lod0, dudv);
	float lod0a = max(d.x, d.y);

	d = calcAlpha(fs_in.pos.xz, grid_size_lod1, dudv);
	float lod1a = max(d.x, d.y);

	d = calcAlpha(fs_in.pos.xz, grid_size_lod2, dudv);
	float lod2a = max(d.x, d.y);
	float lod_fade = fract(LOD);
	vec4 color;
	if(lod2a > 0.0) {
		color = color_thick;
		color.a *= lod2a;
	}else{
		if(lod1a > 0.0) {
			color = mix(color_thick, color_thin, lod_fade);
			color.a *= lod1a;
		}else{
			color = color_thin;
			color.a *= (lod0a * 1.0 - lod_fade);
		}
	}
	float wl = length(fs_in.pos.xz - fs_in.campos) / grid_size;
	float opacity_falloff = 1.0 - clamp(wl, 0.0, 1.0);
	color.a *= opacity_falloff;
	out_color = color;

}
`
)
