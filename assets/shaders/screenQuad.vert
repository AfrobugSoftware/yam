#version 430 core

const vec3 vertices[6] = vec3[6](
    vec3(-1.0, -1.0, 0.5),
    vec3(-1.0, 1.0, 0.5),
    vec3(1.0,1.0,0.5),
    vec3(-1.0, -1.0,0.5),
    vec3(1.0,1.0,0.5),
    vec3(1.0, -1.0, 0.5)
);

out vec2 fragUv

void main() {
    vec4 pos = vec4(vertices[gl_VertexID], 1.0);
    gl_Position = pos;
    fragUv = vec2((pos.x +1) / 2.0, (pos.y + 1) / 2.0);
}