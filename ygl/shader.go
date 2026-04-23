package ygl

import (
	"fmt"
	"os"
	"strings"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

const (
	VERTEX   = gl.VERTEX_SHADER
	FRAGMENT = gl.FRAGMENT_SHADER
)

func CreateShaderFromFile(filename string, shaderType uint32) (uint32, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	return CreateShader(string(source), shaderType)
}

func CreateShader(source string, shaderType uint32) (uint32, error) {
	s := gl.CreateShader(shaderType)
	source = source + "\x00"
	s_source, free := gl.Strs(source)
	gl.ShaderSource(s, 1, s_source, nil)
	free()
	gl.CompileShader(s)
	var s_status int32
	gl.GetShaderiv(s, gl.COMPILE_STATUS, &s_status)
	if s_status != gl.TRUE {
		infoLog := make([]uint8, 2048)
		var length int32
		gl.GetShaderInfoLog(s, int32(len(infoLog)), &length, &infoLog[0])
		var sb strings.Builder
		for _, c := range infoLog {
			sb.WriteByte(byte(c))
		}
		return 0, fmt.Errorf("%d shader failed to compile: %s", shaderType, sb.String())
	}
	return s, nil
}

func CreateProgram(shaders []uint32) (uint32, error) {
	p := gl.CreateProgram()
	for _, s := range shaders {
		gl.AttachShader(p, s)
		gl.DeleteShader(s)
	}
	gl.LinkProgram(p)
	var status int32
	gl.GetProgramiv(p, gl.LINK_STATUS, &status)
	if status != gl.TRUE {
		infoLog := make([]uint8, 2048)
		var length int32
		gl.GetProgramInfoLog(p, int32(len(infoLog)), &length, &infoLog[0])
		var sb strings.Builder
		for _, c := range infoLog {
			sb.WriteByte(byte(c))
		}
		return 0, fmt.Errorf("program failed to link: %s", sb.String())
	}
	return p, nil
}

func SetActiveProgram(p uint32) {
	gl.UseProgram(p)
}

func DestroyProgram(p uint32) {
	gl.DeleteProgram(p)
}

func AssignUniformMat4(p uint32, name string, mat y3d.Mat4) error {
	loc := gl.GetUniformLocation(p, gl.Str(name+"\x00"))
	if loc == -1 {
		return fmt.Errorf("no uniform mat4 with name: %s\n", name)
	}
	gl.UniformMatrix4fv(loc, 1, false, &mat[0])
	return nil
}

func AssignUniformMat4Array(p uint32, name string, count int, mat []float32) error {
	loc := gl.GetUniformLocation(p, gl.Str(name+"\x00"))
	if loc == -1 {
		return fmt.Errorf("no uniform mat4 with name: %s\n", name)
	}
	gl.UniformMatrix4fv(loc, int32(count), false, &mat[0])
	return nil
}

func AssignUniformVec3(p uint32, name string, v y3d.Vec3) error {
	loc := gl.GetUniformLocation(p, gl.Str(name+"\x00"))
	if loc == -1 {
		return fmt.Errorf("no uniform vec3 with name: %s\n", name)
	}
	vs := v.ToSlice()
	gl.Uniform3fv(loc, 1, &vs[0])
	return nil
}
func AssignUniformVec4(p uint32, name string, v y3d.Vec4) error {
	loc := gl.GetUniformLocation(p, gl.Str(name+"\x00"))
	if loc == -1 {
		return fmt.Errorf("no uniform vec3 with name: %s\n", name)
	}
	vs := v.ToSlice()
	gl.Uniform4fv(loc, 1, &vs[0])
	return nil
}

func AssignUniformFloat32(p uint32, name string, f float32) error {
	loc := gl.GetUniformLocation(p, gl.Str(name+"\x00"))
	if loc == -1 {
		return fmt.Errorf("no uniform float with name: %s\n", name)
	}
	gl.Uniform1f(loc, f)
	return nil
}
