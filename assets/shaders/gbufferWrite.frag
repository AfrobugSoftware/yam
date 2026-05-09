#version 430 core

layout (location = 0) out vec3 outDiffuse;
layout (location = 1) out vec3 outWolrdPos;
layout (location = 2) out vec3 outWorldNormal;

in vec3 fragPos;
in vec3 fragNormal;
in vec2 fragUv;

struct Material {
    sampler2D diffuse;
};

uniform Material material;

void main() {
    outDiffuse = texture(material.diffuse, fragUv.xy).xyz;
    outWolrdPos = fragPos;
    outWorldNormal = fragNormal;
}