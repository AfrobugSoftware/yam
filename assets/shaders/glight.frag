#version 430 core
layout (location = 0) out vec4 outColor;

in vec2 fragUv;

struct Light {
    vec3 diffuse;
    vec3 ambient;
    vec3 specular;
    vec3 position;
};

#define MAX_LIGHTS 100
layout(std140, binding = 0) uniform LightSet {
    Light lights[MAX_LIGHTS];
};
layout(location = 3) uniform vec3 cameraPos;
layout(location = 4) uniform int lightCount;

layout(binding = 0) uniform sampler2D udiffuse;
layout(binding = 1) uniform sampler2D uposition;
layout(binding = 2) uniform sampler2D unormal;

vec3 caluatePhongLight(Light light, vec3 fragPos, vec3 fragNormal, float shininess) {
    vec3 ambient = light.ambient;

    vec3 norm           = normalize(fragNormal);
    vec3 lightDir       = normalize(light.position - fragPos);
    float diffuseFactor = max(dot(norm, lightDir), 0.0);
    vec3 diffuse        = light.diffuse * diffuseFactor;

    vec3 camDir          = normalize(cameraPos - fragPos);
    vec3 r               = normalize(reflect(-lightDir, norm));
    float specularFactor = max(dot(r, camDir), 0.0);
    vec3 specular        = pow(specularFactor, shininess) * light.specular;

    return diffuse + ambient;
}

void main () {
    vec3 diffuse  = (texture(udiffuse, fragUv.xy)).xyz;
    vec3 position = (texture(uposition, fragUv.xy)).xyz;
    vec3 normal   = (texture(unormal, fragUv.xy)).xyz;

    vec3 color = vec3(0.0);
    for (int i = 0; i < lightCount; i++) {
        vec3 phong = caluatePhongLight(lights[i], position, normal, 1);
        color += diffuse * phong;
    }
    outColor = vec4(color, 1.0);    
}

