#version 430

out vec4 outColor;

layout(std140, binding= 16) uniform Material {
    vec4 Diffuse;
    vec4 Ambient;
    vec4 Specular;
    vec4 EmissiveShinines;
} material;

struct Light {
    vec4 pos;
    vec4 color;
};

layout(std140, binding=17) uniform Lights {
    Light lights[10];
} l;

void main() {
    outColor = vec4(1.0,0.0,0.0,1.0);
}