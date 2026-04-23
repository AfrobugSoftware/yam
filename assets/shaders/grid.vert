#version 430 core

out vec3 worldPos;

uniform float gridSize = 100.0;
uniform mat4 projView;
uniform vec3 camPos;

const vec3 pos[4] = vec3[4](
    vec3(-1.0,0.0,-1.0),
    vec3(1.0,0.0,-1.0),
    vec3(1.0,0.0,1.0),
    vec3(-1.0,0.0,1.0));

const int indices[6] = int[6](0,2,1,2,0,3);

void main(){
    int index = indices[gl_VertexID];
    vec3 vpos = pos[index] * gridSize;
    vpos.x += camPos.x;
    vpos.y += camPos.y;

    vec4 vpos4 = vec4(vpos, 1.0);
    gl_Position = projView * vpos4;

    worldPos = vpos;
}