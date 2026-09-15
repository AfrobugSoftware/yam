#version 430

layout(location = 0) in vec3 pos;
layout(location = 10) in uint dInx;

layout(std430, binding = 10) buffer MatrixSSBO {
    mat4 world[];
};

layout(location = 0) uniform mat4 viewProj;

void main() {
    gl_Position = viewProj * world[dInx] * vec4(pos, 1.0);
}