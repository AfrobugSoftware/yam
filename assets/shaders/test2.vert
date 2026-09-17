#version 430

layout(location = 0) in vec3 pos;
layout(location = 1) in vec3 normal;
layout(location = 2) in vec2 uv;
layout(location = 10) in uint dInx;

layout(std430, binding = 10) buffer MatrixSSBO {
    mat4 world[];
};

out VS_OUT {
    vec4 frag_pos;
    vec3 normal;
    vec2 uv;
}vs_out;

layout(location = 0) uniform mat4 view_proj;

void main() {
    vs_out.frag_pos = world[dInx] * vec4(pos,1.0);
    vs_out.normal = mat3(world[dInx]) * normal; //mat3(transpose(inverse(world[dInx]))) * normal; this is the actual calculation 
    vs_out.us = uv;


    gl_Position = view_proj * vc_out.frag_pos;
}