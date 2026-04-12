#version 330

layout(location = 0) in vec3 position;
layout(location = 1) in vec3 normal;
layout(location = 2) in vec2 uv;
layout(location = 3) in vec3 jointId;
layout(location = 4) in vec3 jointWeight;

out vec2 fragUv;
out vec3 fragNormal;
out vec3 fragPos;

const int MAX_JOINTS = 100;
const int MAX_WEIGHTS = 3;
uniform mat4 projView;
uniform mat4 world;
uniform mat4 jointTransforms[MAX_JOINTS];


void main() {
    vec4 totalLocalPos = vec4(0.0);
    vec3 totalNormal = vec3(0.0);
    for(int i = 0; i < MAX_WEIGHTS; i++) {
        ivec3 ids = ivec3(jointId);
        vec4 localPostition = jointTransforms[ids[i]] * vec4(position, 1.0);
        totalLocalPos += localPostition * jointWeight[i];

    //joint transforms have no scale, no need to invert/transpose for normal
        vec3 localNormal = mat3(jointTransforms[ids[i]]) * normal; 
        totalNormal += localNormal * jointWeight[i];
    }

    gl_Position = projView * world * totalLocalPos;
    fragNormal = totalNormal;
    fragUv = uv;
    fragPos = totalLocalPos;
}