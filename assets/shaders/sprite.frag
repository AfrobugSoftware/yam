#version 430 core 

layout (location = 0) out vec4 FragColor;

uniform sampler2D tex;
in vec2 TexCoords0;

void main () {
    FragColor = texture2D(tex, TexCoords0.xy);
    if (FragColor == vec4(0.0)) {
        discard;
    }
   // FragColor = vec4(TexCoords0, 0.0,1.0);
}
