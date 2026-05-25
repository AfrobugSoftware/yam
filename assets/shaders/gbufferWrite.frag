#version 430 core

layout (location = 0) out vec3 outDiffuse;
layout (location = 1) out vec3 outWolrdPos;
layout (location = 2) out vec3 outWorldNormal;

in vec3 fragPos;
in vec3 fragNormal;
in vec2 fragUv;

in VS_OUT {
    vec3 fragPos;
    vec3 fragNormal;
    vec2 fragUv;
}fs_in;


layout(binding = 0) uniform sampler2D mdiffuse;

void main() {
    outDiffuse = texture(mdiffuse, fs_in.fragUv.xy).xyz;
    outWolrdPos = fs_in.fragPos;
    outWorldNormal = fs_in.fragNormal;
}