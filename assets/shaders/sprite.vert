#version 430 core

layout (location = 0) in vec2 position;
layout (location = 1) in uint quadId;

#define MAX_QUADS 1000
uniform QuadInfo {
    vec2 BasePos[MAX_QUADS];
    vec2 WidthHeight[MAX_QUADS];
    vec2 TexCoords[MAX_QUADS];
    vec2 TexWidthHeight[MAX_QUADS];
};

out vec2 TexCoords0;

void main() {
    vec3 basePos = vec3(BasePos[quadId], 0.5);
    vec2 widthHeight = position * WidthHeight[quadId];
    vec3 newPosition = basePos +  vec3(widthHeight, 0.0);
    gl_Position = vec4(newPosition, 1.0);

    vec2 baseTexCoords = TexCoords[quadId];
    vec2 texWidthHeight = position * TexWidthHeight[quadId];
    vec2 aTexCoord = baseTexCoords + texWidthHeight;
    TexCoords0 = vec2(aTexCoord.x, 1.0 - aTexCoord.y);
}