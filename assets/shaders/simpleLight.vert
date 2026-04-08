#version 330

layout(location = 0) in vec3 pos;
layout(location = 1) in vec3 normal;
layout(location = 2) in vec2 uv;

out vec2 fragUv;
out vec3 fragNormal;
out vec3 fragPos;

uniform mat4 projView;
uniform mat4 world;

void main() {
    fragUv = uv;
    fragPos = vec3(world * vec4(pos, 1.0));
    fragNormal = mat3(transpose(inverse(world))) * normal; 
    gl_Position = projView * world * vec4(pos, 1.0);
}