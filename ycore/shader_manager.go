package ycore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type ShaderManager struct {
	Shaders map[string]uint32
}

// helpers
func createShader(source string, shaderType uint32) (uint32, error) {
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
		_, err := sb.Write([]byte(infoLog))
		if err != nil {
			return 0, fmt.Errorf("failed to write info log: %v", err)
		}
		return 0, fmt.Errorf("%d shader failed to compile: %s", shaderType, sb.String())
	}
	return s, nil
}

func createProgram(shaders []uint32) (uint32, error) {
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

func createShaderFromFile(filename string, shaderType uint32) (uint32, error) {
	source, err := os.ReadFile(filename)
	if err != nil {
		return 0, err
	}
	return createShader(string(source), shaderType)
}

func NewShaderManager() *ShaderManager {
	return &ShaderManager{
		Shaders: make(map[string]uint32),
	}
}

func (s *ShaderManager) AddFromFile(name string, filename []string, shaderType []uint32) error {
	if len(filename) != len(shaderType) {
		return errors.New("shader type must match shader file names")
	}
	shaders := make([]uint32, len(filename))
	for i, f := range filename {
		sh, err := createShaderFromFile(f, shaderType[i])
		if err != nil {
			return err
		}
		shaders = append(shaders, sh)
	}
	p, err := createProgram(shaders)
	if err != nil {
		return err
	}
	s.Shaders[name] = p
	return nil
}

func (s *ShaderManager) Add(name string, r []io.Reader, shaderType []uint32) error {
	if len(r) != len(shaderType) {
		return errors.New("shader type must match shader file names")
	}
	shaders := make([]uint32, len(r))
	for i, f := range r {
		b, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		sh, err := createShader(string(b), shaderType[i])
		if err != nil {
			return err
		}
		shaders = append(shaders, sh)
	}
	p, err := createProgram(shaders)
	if err != nil {
		return err
	}
	s.Shaders[name] = p
	return nil
}
