#version 330

out vec4 outColor

layout(std140, binding= 16) uniform Material {
    vec4 Diffuse;
    vec4 Ambient;
    vec4 Specular;
    vec4 EmissiveShinines;
};

void main() {
    outColor = Diffuse;
}