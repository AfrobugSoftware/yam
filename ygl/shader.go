package ygl

import (
	"fmt"
	"os"
	"strings"

	gl "github.com/chsc/gogl/gl33"
)

const (
	VERTEX   = gl.VERTEX_SHADER
	FRAGMENT = gl.FRAGMENT_SHADER
)

func CreateShaderFromFile(filename string, shaderType gl.Enum) (gl.Uint, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	return CreateShader(string(source), shaderType)
}

func CreateShader(source string, shaderType gl.Enum) (gl.Uint, error) {
	s := gl.CreateShader(shaderType)
	s_source := gl.GLString(source)
	gl.ShaderSource(s, 1, &s_source, nil)
	gl.CompileShader(s)
	var s_status gl.Int
	gl.GetShaderiv(s, gl.COMPILE_STATUS, &s_status)
	if s_status != gl.TRUE {
		infoLog := make([]gl.Char, 2048)
		var length gl.Sizei
		gl.GetShaderInfoLog(s, gl.Sizei(len(infoLog)), &length, &infoLog[0])
		var sb strings.Builder
		for _, c := range infoLog {
			sb.WriteByte(byte(c))
		}
		return 0, fmt.Errorf("shader failed to compile: %s", sb.String())
	}
	return s, nil
}

func CreateProgram(shaders []gl.Uint) (gl.Uint, error) {
	p := gl.CreateProgram()
	for _, s := range shaders {
		gl.AttachShader(p, s)
		gl.DeleteShader(s)
	}
	gl.LinkProgram(p)
	var status gl.Int
	gl.GetProgramiv(p, gl.LINK_STATUS, &status)
	if status != gl.TRUE {
		infoLog := make([]gl.Char, 2048)
		var length gl.Sizei
		gl.GetProgramInfoLog(p, gl.Sizei(len(infoLog)), &length, &infoLog[0])
		var sb strings.Builder
		for _, c := range infoLog {
			sb.WriteByte(byte(c))
		}
		return 0, fmt.Errorf("program failed to link: %s", sb.String())
	}
	return p, nil
}

func SetActive(p gl.Uint) {
	gl.UseProgram(p)
}

func DestroyProgram(p gl.Uint) {
	gl.DeleteProgram(p)
}
