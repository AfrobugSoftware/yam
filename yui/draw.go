package yui

import (
	"fmt"
	"strings"
	"yam/y3d"

	"github.com/go-gl/gl/v4.3-core/gl"
)

var (
	arrayBuffer  *QuadArray
	arrayProgram uint32
)

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

func Setup() error {
	v, err := createShader(vertShader, gl.VERTEX_SHADER)
	if err != nil {
		return err
	}
	f, err := createShader(fragShader, gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}
	p, err := createProgram([]uint32{v, f})
	if err != nil {
		return err
	}
	arrayBuffer = CreateQuadArray(p)
	arrayProgram = p
	return nil
}

func DrawText(baseInstance int, ui *UIElement) int {
	if ui.Font == nil {
		return baseInstance
	}
	count := baseInstance
	x := ui.Position.X
	y := ui.Position.Y
	for _, cand := range ui.Lines {
		for _, c := range cand {
			char, ok := ui.Font.GetCharacter(c)
			if c == ' ' {
				x += char.Advance
			}
			if !ok {
				//or render missing charater glyph
				continue
			}
			arrayBuffer.SetQuad(count,
				x,
				y,
				char.Size.X,
				(char.Size.Y),
				0,
				0,
				1, 1,
				float32(ui.Color[0])/255.0,
				float32(ui.Color[1])/255.0,
				float32(ui.Color[2])/255.0,
				float32(char.Idx), ui.Radius, ui.RenderFlag|UI_TEXT, ui.Tex)
			count++
			x = x + char.Advance + float32(ui.Font.Kern)
		}
		y += ui.Font.VerticalSpacing + float32(ui.Font.height)
	}
	return count
}

func Draw(root *UIElement, proj y3d.Mat4) {
	flatten(root)
	var font *Font
	gl.UseProgram(arrayProgram)
	gl.UniformMatrix4fv(1, 1, false, &proj[0])
	err := arrayBuffer.OpenArray()
	if err != nil {
		panic(err)
	}
	i := 0
	for _, ui := range flat {
		if len(ui.Lines) > 0 {
			i = DrawText(i, ui)
			font = ui.Font
		} else {
			arrayBuffer.SetQuad(i,
				ui.Position.X,
				ui.Position.Y,
				ui.Size.X,
				ui.Size.Y,
				ui.TextureBase[0],
				ui.TextureBase[1],
				1, 1,
				float32(ui.Color[0])/255.0,
				float32(ui.Color[1])/255.0,
				float32(ui.Color[2])/255.0,
				float32(ui.Color[3])/255.0, ui.Radius, ui.RenderFlag, ui.Tex)
			i++
		}
	}
	arrayBuffer.CloseArray()
	arrayBuffer.Draw(i, font)
	gl.UseProgram(0)
}
