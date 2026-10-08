package ygame

import (
	"log"
	"unsafe"
	"yam/ycore"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

const vertexShader = `#version 430 core
layout(location = 0) in vec2 Position;
layout(location = 1) in vec2 UV;
layout(location = 2) in vec4 Color;

layout(location = 1) uniform mat4 ProjMtx;
out vec2 Frag_UV;
out vec4 Frag_Color;
void main() {
    Frag_UV = UV;
    Frag_Color = Color;
    gl_Position = ProjMtx * vec4(Position, 0.0, 1.0);
}`

const fragmentShader = `#version 430 core
in vec2 Frag_UV;
in vec4 Frag_Color;
layout(location = 0) out vec4 Out_Color;

layout(location = 1) uniform sampler2D Texture;
void main() {
    Out_Color = Frag_Color * texture(Texture, Frag_UV);
}`

type renderer struct {
	program uint32
	vao     uint32
	vbo     uint32
	ebo     uint32
	fontTex uint32
}

func newRenderer(io *imgui.IO, s *ycore.ShaderManager) *renderer {
	r := &renderer{}
	err := s.Add("imgui", []string{
		vertexShader, fragmentShader,
	},
		[]uint32{
			gl.VERTEX_SHADER,
			gl.FRAGMENT_SHADER,
		},
	)
	if err != nil {
		panic(err)
	}
	p := s.Shaders["imgui"]
	r.program = p

	gl.GenVertexArrays(1, &r.vao)
	gl.GenBuffers(1, &r.vbo)
	gl.GenBuffers(1, &r.ebo)

	vtxSize, posOff, uvOff, colOff := imgui.VertexBufferLayout()
	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.ebo)
	gl.EnableVertexAttribArray(0)
	gl.EnableVertexAttribArray(1)
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, int32(vtxSize), uintptr(posOff))
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, int32(vtxSize), uintptr(uvOff))
	gl.VertexAttribPointerWithOffset(2, 4, gl.UNSIGNED_BYTE, true, int32(vtxSize), uintptr(colOff))
	gl.BindVertexArray(0)

	texData := io.Fonts().TexData()
	gl.GenTextures(1, &r.fontTex)
	gl.BindTexture(gl.TEXTURE_2D, r.fontTex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, texData.Width(),
		texData.Height(),
		0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(texData))

	return r
}

func (r *renderer) render(dd *imgui.DrawData) {
	disp, scale := dd.DisplaySize(), dd.FramebufferScale()
	fbW, fbH := disp.X*scale.X, disp.Y*scale.Y
	if fbW <= 0 || fbH <= 0 {
		return
	}
	pos := dd.DisplayPos()

	gl.Enable(gl.BLEND)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.SCISSOR_TEST)

	L, R := pos.X, pos.X+disp.X
	T, B := pos.Y, pos.Y+disp.Y
	proj := [16]float32{
		2 / (R - L), 0, 0, 0,
		0, 2 / (T - B), 0, 0,
		0, 0, -1, 0,
		(R + L) / (L - R), (T + B) / (B - T), 0, 1,
	}

	gl.UseProgram(r.program)
	gl.UniformMatrix4fv(0, 1, false, &proj[0])
	gl.Uniform1i(1, 0)
	gl.BindVertexArray(r.vao)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, r.fontTex)

	idxSize := imgui.IndexBufferLayout()
	idxType := uint32(gl.UNSIGNED_SHORT)
	if idxSize == 4 {
		idxType = gl.UNSIGNED_INT
	}

	for _, list := range dd.CommandLists() {
		vb, vbSize := list.GetVertexBuffer()
		ib, ibSize := list.GetIndexBuffer()
		gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, vbSize, vb, gl.STREAM_DRAW)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, ibSize, ib, gl.STREAM_DRAW)

		for _, cmd := range list.Commands() {
			clip := cmd.ClipRect()
			x1 := (clip.X - pos.X) * scale.X
			y1 := (clip.Y - pos.Y) * scale.Y
			x2 := (clip.Z - pos.X) * scale.X
			y2 := (clip.W - pos.Y) * scale.Y
			if x2 <= x1 || y2 <= y1 {
				continue
			}
			gl.Scissor(int32(x1), int32(fbH-y2), int32(x2-x1), int32(y2-y1))
			gl.DrawElementsBaseVertexWithOffset(
				gl.TRIANGLES,
				int32(cmd.ElemCount()),
				idxType,
				uintptr(int(cmd.IdxOffset())*idxSize),
				int32(cmd.VtxOffset()),
			)
		}
	}

	gl.Disable(gl.SCISSOR_TEST)
	gl.BindVertexArray(0)
}

func (r *renderer) destroy() {
	gl.DeleteTextures(1, &r.fontTex)
	gl.DeleteBuffers(1, &r.vbo)
	gl.DeleteBuffers(1, &r.ebo)
	gl.DeleteVertexArrays(1, &r.vao)
	gl.DeleteProgram(r.program)
}

func mouseButton(b uint8) (int32, bool) {
	switch b {
	case sdl.BUTTON_LEFT:
		return 0, true
	case sdl.BUTTON_RIGHT:
		return 1, true
	case sdl.BUTTON_MIDDLE:
		return 2, true
	}
	return 0, false
}

var keyMap = map[int]imgui.Key{
	sdl.SCANCODE_TAB:       imgui.KeyTab,
	sdl.SCANCODE_LEFT:      imgui.KeyLeftArrow,
	sdl.SCANCODE_RIGHT:     imgui.KeyRightArrow,
	sdl.SCANCODE_UP:        imgui.KeyUpArrow,
	sdl.SCANCODE_DOWN:      imgui.KeyDownArrow,
	sdl.SCANCODE_PAGEUP:    imgui.KeyPageUp,
	sdl.SCANCODE_PAGEDOWN:  imgui.KeyPageDown,
	sdl.SCANCODE_HOME:      imgui.KeyHome,
	sdl.SCANCODE_END:       imgui.KeyEnd,
	sdl.SCANCODE_INSERT:    imgui.KeyInsert,
	sdl.SCANCODE_DELETE:    imgui.KeyDelete,
	sdl.SCANCODE_BACKSPACE: imgui.KeyBackspace,
	sdl.SCANCODE_SPACE:     imgui.KeySpace,
	sdl.SCANCODE_RETURN:    imgui.KeyEnter,
	sdl.SCANCODE_ESCAPE:    imgui.KeyEscape,
	sdl.SCANCODE_A:         imgui.KeyA,
	sdl.SCANCODE_C:         imgui.KeyC,
	sdl.SCANCODE_V:         imgui.KeyV,
	sdl.SCANCODE_X:         imgui.KeyX,
	sdl.SCANCODE_Y:         imgui.KeyY,
	sdl.SCANCODE_Z:         imgui.KeyZ,
}

type Editor struct {
	Input              *ycore.InputManager
	RenderManager      *ycore.RenderManager
	VertexCacheManager *ycore.VertexCacheManager
	ShaderManager      *ycore.ShaderManager
	window             *sdl.Window
	IO                 *imgui.IO
	R                  *renderer
}

func NewEditor(input *ycore.InputManager,
	rm *ycore.RenderManager,
	vcm *ycore.VertexCacheManager,
	shader *ycore.ShaderManager,
	w *sdl.Window) *Editor {
	imgui.CreateContext()
	io := imgui.CurrentIO()
	io.SetBackendFlags(io.BackendFlags() | imgui.BackendFlagsRendererHasVtxOffset)
	e := &Editor{
		Input:              input,
		RenderManager:      rm,
		VertexCacheManager: vcm,
		window:             w,
		ShaderManager:      shader,
		IO:                 imgui.CurrentIO(),
		R:                  newRenderer(io, shader),
	}
	return e
}

func (e *Editor) UpdateInput() {
	//pressed or held ?
	e.IO.AddMousePosEvent(e.Input.MousePosition.X, e.Input.MousePosition.Y)
	e.IO.AddMouseButtonEvent(1, e.Input.GetMouseButtonState(1) == ycore.BUTTON_PRESSED)
	e.IO.AddMouseButtonEvent(2, e.Input.GetMouseButtonState(2) == ycore.BUTTON_PRESSED)
	e.IO.AddMouseButtonEvent(3, e.Input.GetMouseButtonState(3) == ycore.BUTTON_PRESSED)

	e.IO.AddKeyEvent(imgui.ModCtrl, e.Input.GetKeyStateMod(ycore.CTRL) == ycore.BUTTON_PRESSED)
	e.IO.AddKeyEvent(imgui.ModShift, e.Input.GetKeyStateMod(ycore.SHIFT) == ycore.BUTTON_PRESSED)
	e.IO.AddKeyEvent(imgui.ModAlt, e.Input.GetKeyStateMod(ycore.ALT) == ycore.BUTTON_PRESSED)
	e.IO.AddKeyEvent(imgui.ModSuper, e.Input.GetKeyStateMod(ycore.SUPER) == ycore.BUTTON_PRESSED)
	for p, k := range e.Input.CurKeyState {
		ik, ok := keyMap[p]
		if ok {
			e.IO.AddKeyEvent(ik, k == ycore.BUTTON_PRESSED)
		}
	}
	e.IO.AddInputCharactersUTF8(e.Input.TextInput)
}

func (e *Editor) Frame(dt float32) {
	e.UpdateInput()

	w, h := e.window.GetSize()
	fbW, fbH := e.window.GLGetDrawableSize()
	e.IO.SetDisplaySize(imgui.Vec2{X: float32(w), Y: float32(h)})
	if w > 0 && h > 0 {
		e.IO.SetDisplayFramebufferScale(imgui.Vec2{
			X: float32(fbW) / float32(w),
			Y: float32(fbH) / float32(h),
		})
	}
	e.IO.SetDeltaTime(dt)

	//test
	imgui.NewFrame()

	imgui.SetNextWindowPosV(imgui.Vec2{X: 60, Y: 60}, imgui.CondFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 320, Y: 140}, imgui.CondFirstUseEver)
	imgui.Begin("Hello, cimgui-go")
	imgui.Text("SDL2 + OpenGL 4.3 core")
	if imgui.Button("Click me") {
		log.Println("clicking")
	}
	imgui.SameLine()
	imgui.Text("This is a text line")
	imgui.End()

	imgui.Render()
	e.R.render(imgui.CurrentDrawData())
}

func (e *Editor) Destroy() {
	imgui.DestroyContext()
}
