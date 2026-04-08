#version 330
out vec4 color;

in vec2 fragUv;
in vec3 fragNormal;
in vec3 fragPos;

uniform vec3 cameraPos;

struct Material {
    vec3 ambient;
    vec3 diffuse;
    vec3 specular;
    float shininess;
};

struct Light {
    vec3 diffuse;
    vec3 ambient;
    vec3 specular;
    vec3 position;
};

uniform Material material;
uniform Light light;

vec3 caluatePhongLight() {
    vec3 ambient = light.ambient * material.ambient;

    vec3 norm = normalize(fragNormal);
    vec3 lightDir = normalize(light.position - fragPos);
    float diffuseFactor = max(dot(norm, lightDir), 0.0);
    vec3 diffuse = light.diffuse * (material.diffuse * diffuseFactor);

    vec3 camDir = normalize(cameraPos - fragPos);
    vec3 r = reflect(-lightDir, norm);
    float specularFactor = max(dot(r, camDir), 0.0);
    vec3 specular = (material.specular *  pow(specularFactor, material.shininess)) * light.specular;

    return ambient + diffuse + specular;
}


void main() {
    vec3 lighting = caluatePhongLight();
    color = vec4(lighting, 1.0);
}