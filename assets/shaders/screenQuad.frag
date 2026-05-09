#version 430 core
layout (location = 0) out vec4 OutColor;

in vec2 fragUv;

void main () {
    OutColor = vec4(fragUv, 1.0, 1.0);
}