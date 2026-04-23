#version 430 core
layout(location = 0) out vec4 FragColor;

in vec3 worldPos;

uniform float gridCellSize = 0.025;
uniform float gridMinPixelsBetweenCells = 2.0;
uniform vec4  gridColorThin = vec4(0.5,0.5,0.5,1.0);
uniform vec4  gridColorThick = vec4(0.0,0.0,0.0,1.0);
uniform float gridSize = 100.0;

void main() {
    vec2 dvy = vec2(dFdx(worldPos.z), dFdy(worldPos.z));
    vec2 dvx = vec2(dFdx(worldPos.x), dFdy(worldPos.x));

    float ly = length(dvy);
    float lx = length(dvx);

    vec2 dudv = vec2(lx,ly);
    float l = length(dudv);
    float LOD = max(0.0, log10(l * gridMinPixelsBetweenCells / gridCellSize) + 1);
    float gridCellSizeLod0 = gridCellSize * pow(10.0, floor(LOD));
    float gridCellSizeLod1 = gridCellSizeLod0 * 10.0;
    float gridCellSizeLod2 = gridCellSizeLod1 * 10.0;
    dudv *= 4.0;

    vec2 modDivDudv = mod(worldPos.xz, gridCellSizeLod0) / dudv;
    float lod0a = max(0.0,vec2(1.0) - abs(stav(modDivDudv) * 2.0 - vec2(1.0)));
    vec2 modDivDudv = mod(worldPos.xz, gridCellSizeLod1) / dudv;
    float lod1a = max(0.0,vec2(1.0) - abs(stav(modDivDudv) * 2.0 - vec2(1.0)));
    vec2 modDivDudv = mod(worldPos.xz, gridCellSizeLod2) / dudv;
    float lod2a = max(0.0,vec2(1.0) - abs(stav(modDivDudv) * 2.0 - vec2(1.0)));

    float LODFade = fract(LOD);
    vec4 color;
    if(lod2a > 0.0) {
        color = gridColorThick;
        color.a += lod2a;
    }else {
        if(lod1a > 0.0) {
            color = mix(gridColorThick, gridColorThin, LODFade);
            color.a *= lod1a;
        }else {
            color = gridColorThick;
            color.a *= (lod0a * ( 1.0 -  LODFade));
        }
    }
    FragColor = color;
}