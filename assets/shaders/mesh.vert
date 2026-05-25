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

//layout(std140, binding = 1) uniform Transform {
  //  mat4 world[];
//};


void main() {
    vs_out.fragUv = uv;
    vs_out.fragPos = vec3(world * vec4(pos, 1.0));
    vs_out.fragNormal = mat3(inverse(transpose(world))) * normal; 
   
    gl_Position = projView * world * vec4(pos, 1.0);
}