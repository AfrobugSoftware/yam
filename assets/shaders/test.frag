#version 430

out vec4 outColor;

layout(std140, binding= 2) uniform Material {
    vec4 diffuse;
    vec4 ambient;
    vec4 specular;
    vec4 emissiveShinines;
} material;

uniform sampler2D shadawMap; 

void main() {
    outColor = vec4(1.0,0.0,0.0,1.0);
}