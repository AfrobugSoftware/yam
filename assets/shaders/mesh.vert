#version 430 core
layout(location = 0) in vec3 pos;
layout(location = 1) in vec3 normal;
layout(location = 2) in vec2 uv;

out VS_OUT {
    vec3 fragPos;
    vec3 fragNormal;
    vec2 fragUv;
}vs_out;

layout (location = 0) uniform mat4 projView; 
layout (location = 1) uniform mat4 world;
layout (location = 2) uniform mat3 normalMat;


void main() {
    vs_out.fragUv = uv;
    vs_out.fragPos = vec3(world * vec4(pos, 1.0));
    vs_out.fragNormal = normalMat * normal; 
   
    gl_Position = projView * world * vec4(pos, 1.0);
}