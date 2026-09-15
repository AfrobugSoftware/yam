package ycore

import (
	"fmt"
	"strings"
	"unsafe"
	"yam/y3d"
	"yam/ygl"

	"github.com/go-gl/gl/v4.3-core/gl"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	MODE_2D             = 0
	MODE_3D_ORTHO       = 1
	MODE_3D_PERSPECTIVE = 2
)

const (
	MAX_LIGHT                = 10
	VERTEX_ATTRIBUTE_BINDING = 0
	DRAW_INDEX_BINDING       = 10
	MATERIAL_SSBO_BINDING    = 16
	LIGHT_BINDING            = 17
	WORLD_MATRIX_BINDING     = 18
)

type RenderManager struct {
	Context        sdl.GLContext
	Window         *sdl.Window
	ClearColor     y3d.Vec4
	PixelDepth     uint8
	DoubleBuffer   bool
	MinorVersion   int
	MajorVersion   int
	ViewPort       [4]y3d.Rect
	View2D         y3d.Mat4
	View3D         y3d.Mat4
	Proj2D         y3d.Mat4
	ProjP          [4]y3d.Mat4
	ProjO          [4]y3d.Mat4
	ViewProj       y3d.Mat4
	WorldViewProj  y3d.Mat4
	Near           float32
	Far            float32
	Width          int
	Height         int
	Fov            [4]float32
	AspectRatio    [4]float32
	Stage          int
	Mode           int
	SkinManager    *SkinManager
	VertextManager *VertexCacheManager
	ShaderManager  *ShaderManager
	fbo            uint32
	ColorBuffers   []uint32
	DepthBuffer    uint32
	RenderStates   []RenderState
	DrawMode       uint32
	ActiveProgram  uint32
	Root           SpatialInterface
	Lights         []ygl.Light
	LightUBO       uint32
	ActiveLights   int
}

func NewRenderManager(window *sdl.Window, width, height int) *RenderManager {
	rm := &RenderManager{
		Window: window,
		ClearColor: y3d.Vec4{
			X: 0,
			Y: 0,
			Z: 0,
			W: 1,
		},
		View2D: y3d.Identity,
		View3D: y3d.Identity,
		Proj2D: y3d.Identity,
		ProjO: [4]y3d.Mat4{
			y3d.Identity,
			y3d.Identity,
			y3d.Identity,
			y3d.Identity,
		},
		ProjP: [4]y3d.Mat4{
			y3d.Identity,
			y3d.Identity,
			y3d.Identity,
			y3d.Identity,
		},
		Mode:     MODE_3D_PERSPECTIVE,
		DrawMode: gl.TRIANGLES,
		Stage:    -1,
	}
	context, err := window.GLCreateContext()
	if err != nil {
		panic(err)
	}
	rm.Context = context
	if err = gl.Init(); err != nil {
		panic(err)
	}
	gl.ClearColor(rm.ClearColor.X,
		rm.ClearColor.Y,
		rm.ClearColor.Z,
		rm.ClearColor.W)
	rm.SkinManager = NewSkinManager()
	rm.VertextManager = NewVertexCacheManager(rm,
		10000, 10000*3, 10000, 10000)
	rm.ShaderManager = NewShaderManager()
	rm.CreateDefaultShader()
	rm.SetClippingPlanes(0.1, 1000.0)

	rm.InitStage(0, y3d.Rect{
		X:      0,
		Y:      0,
		Height: height,
		Width:  width,
	},
		float32(y3d.ToRadians(45)),
	)
	rm.SetStage(MODE_3D_PERSPECTIVE, 0) //SET TO THE 0TH stage

	gl.CreateBuffers(1, &rm.LightUBO)
	gl.NamedBufferStorage(rm.LightUBO, int(112*MAX_LIGHT), nil, gl.DYNAMIC_STORAGE_BIT)
	return rm
}

func (r *RenderManager) CreateFrameBuffer() {
	gl.GenFramebuffers(1, &r.fbo)
	r.ColorBuffers = make([]uint32, 0)
}

func (r *RenderManager) DrawBuffers() {
	drawbuffer := make([]uint32, len(r.ColorBuffers))
	for i := range r.ColorBuffers {
		drawbuffer[i] = uint32(gl.COLOR_ATTACHMENT0 + i)
	}
	gl.DrawBuffers(int32(len(drawbuffer)), &drawbuffer[0])
}

func (r *RenderManager) CheckComplete() bool {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	return status == gl.FRAMEBUFFER_COMPLETE
}

func (r *RenderManager) DestroyFrameBuffer() {
	if r.fbo != 0 {
		gl.DeleteFramebuffers(1, &r.fbo)
		r.fbo = 0
	}
	if r.DepthBuffer != 0 {
		gl.DeleteRenderbuffers(1, &r.DepthBuffer)
		r.DepthBuffer = 0
	}
	for i := range r.ColorBuffers {
		gl.DeleteTextures(1, &r.ColorBuffers[i])
		clear(r.ColorBuffers)
	}
	r.ColorBuffers = nil
}

func (r *RenderManager) CreateColorBufferTexture() {
	var textureColorbuffer uint32
	gl.GenTextures(1, &textureColorbuffer)
	gl.BindTexture(gl.TEXTURE_2D, textureColorbuffer)
	gl.TexImage2D(gl.TEXTURE_2D,
		0,
		gl.RGB,
		int32(r.ViewPort[r.Stage].Width),
		int32(r.ViewPort[r.Stage].Height),
		0, gl.RGB, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	attachment := gl.COLOR_ATTACHMENT0 + len(r.ColorBuffers)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, uint32(attachment), gl.TEXTURE_2D, textureColorbuffer, 0)
	r.ColorBuffers = append(r.ColorBuffers, textureColorbuffer)
	gl.BindTexture(gl.TEXTURE_2D, 0)

}

func (r *RenderManager) CreateDepthTexture() {
	var tex uint32
	gl.GenTextures(1, &tex)
	gl.BindTexture(gl.TEXTURE_2D, tex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH24_STENCIL8,
		int32(r.ViewPort[r.Stage].Width),
		int32(r.ViewPort[r.Stage].Height),
		0,
		gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.TEXTURE_2D, tex, 0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	r.DepthBuffer = tex
}

func (r *RenderManager) BindFrameBuffer() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
}

func (r *RenderManager) UnbindFrameBuffer() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (r *RenderManager) GetFrustum() [6]y3d.Plane {
	var planes [6]y3d.Plane
	//left
	planes[0] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] + r.ViewProj[0]),
			Y: -(r.ViewProj[13] + r.ViewProj[1]),
			Z: -(r.ViewProj[14] + r.ViewProj[2]),
		},
		D: -(r.ViewProj[15] + r.ViewProj[3]),
	}
	//right
	planes[1] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] - r.ViewProj[0]),
			Y: -(r.ViewProj[13] - r.ViewProj[1]),
			Z: -(r.ViewProj[14] - r.ViewProj[2]),
		},
		D: -(r.ViewProj[15] - r.ViewProj[3]),
	}
	//top
	planes[2] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] - r.ViewProj[4]),
			Y: -(r.ViewProj[13] - r.ViewProj[5]),
			Z: -(r.ViewProj[14] - r.ViewProj[6]),
		},
		D: -(r.ViewProj[15] - r.ViewProj[7]),
	}
	//bottom
	planes[3] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] + r.ViewProj[4]),
			Y: -(r.ViewProj[13] + r.ViewProj[5]),
			Z: -(r.ViewProj[14] + r.ViewProj[6]),
		},
		D: -(r.ViewProj[15] + r.ViewProj[7]),
	}
	//near
	planes[4] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[8]),
			Y: -(r.ViewProj[9]),
			Z: -(r.ViewProj[10]),
		},
		D: -(r.ViewProj[11]),
	}
	//bottom
	planes[5] = y3d.Plane{
		N: y3d.Vec3{
			X: -(r.ViewProj[12] - r.ViewProj[8]),
			Y: -(r.ViewProj[13] - r.ViewProj[9]),
			Z: -(r.ViewProj[14] - r.ViewProj[10]),
		},
		D: -(r.ViewProj[15] - r.ViewProj[11]),
	}
	for i := range planes {
		l := planes[i].N.Length()
		planes[i].D /= l
		planes[i].N = y3d.SDiv(planes[i].N, l)
	}
	return planes
}

func (r *RenderManager) SetClippingPlanes(near, far float32) {
	if near > far {
		near = far
		far = near + 1
	}
	if near <= 0.0 {
		near = 0.01
	}
	if far <= 1.0 {
		far = 1.00
	}
	r.Near = near
	r.Far = far
	r.Prepare2D()

	Q := 1.0 / (r.Far - r.Near)
	X := -Q * r.Near
	r.ProjO[0][10] = Q
	r.ProjO[1][10] = Q
	r.ProjO[2][10] = Q
	r.ProjO[3][10] = Q
	r.ProjO[0][14] = X
	r.ProjO[1][14] = X
	r.ProjO[2][14] = X
	r.ProjO[3][14] = X

	Q *= r.Far
	X = -Q * r.Near
	r.ProjP[0][10] = Q
	r.ProjP[1][10] = Q
	r.ProjP[2][10] = Q
	r.ProjP[3][10] = Q
	r.ProjP[0][14] = X
	r.ProjP[1][14] = X
	r.ProjP[2][14] = X
	r.ProjP[3][14] = X
}

func (r *RenderManager) Prepare2D() {
	r.View2D = y3d.Identity
	r.Proj2D[0] = 2.0 / float32(r.Width)
	r.Proj2D[5] = 2.0 / float32(r.Height)
	r.Proj2D[10] = 1.0 / (r.Far - r.Near)
	r.Proj2D[11] = -r.Near * (1.0 / (r.Far - r.Near))
	r.Proj2D[15] = 1.0
	//flip the y axis for 2D rendering``
	r.View2D[5] = -1.0
	r.View2D[12] = -float32(r.Width) + float32(r.Width)*0.5
	r.View2D[13] = float32(r.Height) - float32(r.Height)*0.5
	r.View2D[14] = r.Near + 0.1
}

func (r *RenderManager) CalcViewProj() {
	var a, b *y3d.Mat4
	switch r.Mode {
	case MODE_2D:
		a = &r.Proj2D
		b = &r.View2D
	case MODE_3D_ORTHO:
		a = &r.ProjO[r.Stage]
		b = &r.View3D
	case MODE_3D_PERSPECTIVE:
		a = &r.ProjP[r.Stage]
		b = &r.View3D
	default:
		a = &y3d.Identity
		b = &y3d.Identity
	}
	r.ViewProj = (*a).Mul(*b)
}
func (r *RenderManager) SetStage(mode int, stage int) {
	if stage < 0 || stage >= 4 {
		stage = 0
	}
	if r.Mode != mode || r.Stage != stage {
		r.Mode = mode
		r.Stage = stage
	}
	gl.Viewport(int32(r.ViewPort[stage].X),
		int32(r.ViewPort[stage].Y),
		int32(r.ViewPort[stage].Width),
		int32(r.ViewPort[stage].Height))
	r.CalcViewProj()
}

func (r *RenderManager) InitStage(stage int, viewport y3d.Rect, fov float32) {
	if stage < 0 || stage >= 4 {
		stage = 0
	}
	r.ViewPort[stage] = viewport
	r.Fov[stage] = fov
	r.AspectRatio[stage] = float32(viewport.Width) / float32(viewport.Height)
	r.ProjP[stage] = y3d.Perspective(fov, r.AspectRatio[stage], r.Near, r.Far)
	r.ProjO[stage] = y3d.Identity

	r.ProjO[stage][0] = 2 / float32(viewport.Width)
	r.ProjO[stage][5] = 2 / float32(viewport.Height)
	r.ProjO[stage][10] = 1.0 / (r.Far - r.Near)
	r.ProjO[stage][11] = -r.Near * (1.0 / (r.Far - r.Near))
	r.ProjO[stage][15] = 1.0
}

func (r *RenderManager) Transfrom3Dto2D(pos y3d.Vec3) y3d.Vec2 {
	var width, height float32
	if r.Mode == MODE_2D {
		width, height = float32(r.Width), float32(r.Height)
	} else {
		width, height = float32(r.ViewPort[r.Stage].Width), float32(r.ViewPort[r.Stage].Height)
	}
	clipX := width / 2
	clipY := height / 2

	fXp := r.ViewProj.MulVec4(y3d.Vec4{X: pos.X, Y: pos.Y, Z: pos.Z, W: 1.0})
	Winv := 1.0 / fXp.W
	fXp.X *= Winv
	fXp.Y *= Winv
	fXp.Z *= Winv
	fXp.W = 1.0
	return y3d.Vec2{
		X: (1.0 + fXp.X) * clipX,
		Y: (1.0 + fXp.Y) * clipY,
	}
}

func (r *RenderManager) Transfrom2Dto3D(pos y3d.Vec2) (y3d.Vec3, y3d.Vec3) {
	var width, height float32
	if r.Mode == MODE_2D {
		width, height = float32(r.Width), float32(r.Height)
	} else {
		width, height = float32(r.ViewPort[r.Stage].Width), float32(r.ViewPort[r.Stage].Height)
	}
	vcS := y3d.Vec3{
		X: (((pos.X * 2.0) / width) - 1.0) / r.ProjP[r.Stage][0],
		Y: (((pos.Y * 2.0) / height) - 1.0) / r.ProjP[r.Stage][5],
		Z: 1.0,
	}
	vcS.Y = -vcS.Y
	InvView := r.View3D
	(&InvView).Invert()
	var dir, origin y3d.Vec3
	dir = InvView.MulVec3(vcS)
	origin = y3d.Vec3{
		X: InvView[12],
		Y: InvView[13],
		Z: InvView[14],
	}
	dir = y3d.Normalize(dir)
	return origin, dir
}

func (r *RenderManager) Destroy() {
	r.DestroyFrameBuffer()
	if r.SkinManager != nil {
		r.SkinManager.Destroy()
	}
	if r.VertextManager != nil {
		r.VertextManager.Destory()
	}

	sdl.GLDeleteContext(r.Context)
}

func (r *RenderManager) Clear() {
	gl.ClearColor(r.ClearColor.X, r.ClearColor.Y, r.ClearColor.Z, r.ClearColor.W)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
}

func (r *RenderManager) SetClearColor(color y3d.Vec4) {
	r.ClearColor = color
	gl.ClearColor(r.ClearColor.X, r.ClearColor.Y, r.ClearColor.Z, r.ClearColor.W)
}

func (r *RenderManager) Present() {
	r.Window.GLSwap()
}

func (r *RenderManager) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Major version: %d\n", r.MajorVersion)
	fmt.Fprintf(&b, "Pixel depth: %d\n", r.PixelDepth)
	fmt.Fprintf(&b, "Doubled buffer: %v\n", r.DoubleBuffer)
	fmt.Fprintf(&b, "Height: %d\n", r.Height)
	fmt.Fprintf(&b, "Width:  %d\n", r.Width)
	fmt.Fprintf(&b, "Far: %.2f\n", r.Far)
	fmt.Fprintf(&b, "Near: %.2f\n", r.Near)
	fmt.Fprintf(&b, "FOV: %.2f\n", r.Fov)

	return b.String()
}

func (r *RenderManager) CreateDefaultShader() {
	err := r.ShaderManager.AddFromFile("default",
		[]string{"assets/shaders/test.vert", "assets/shaders/test.frag"},
		[]uint32{gl.VERTEX_SHADER, gl.FRAGMENT_SHADER})
	if err != nil {
		panic(err)
	}
	p, ok := r.ShaderManager.Shaders["default"]
	if ok {
		r.ActiveProgram = p
	}
}

// hmmm
func (r *RenderManager) Render() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	gl.UseProgram(r.ActiveProgram)
	gl.UniformMatrix4fv(0, 1, false, &r.ViewProj[0])

	if r.ActiveLights > 0 {
		gl.NamedBufferSubData(r.LightUBO, 0, int(unsafe.Sizeof(ygl.Light{})*uintptr(MAX_LIGHT)),
			gl.Ptr(r.Lights))
		gl.BindBufferBase(gl.UNIFORM_BUFFER, LIGHT_BINDING, r.LightUBO)
	}
	//might be mvoved to the scene_manager.go
	if r.Root != nil {
		r.Root.Draw(r)
	}
	r.Window.GLSwap()
}

func (r *RenderManager) RenderToGBuffer() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT | gl.STENCIL_BUFFER_BIT)
	gl.UseProgram(r.ActiveProgram)
	gl.UniformMatrix4fv(0, 1, false, &r.ViewProj[0])

	if r.ActiveLights > 0 {
		gl.NamedBufferSubData(r.LightUBO, 0, int(unsafe.Sizeof(ygl.Light{})*uintptr(MAX_LIGHT)),
			gl.Ptr(r.Lights))
		gl.BindBufferBase(gl.UNIFORM_BUFFER, LIGHT_BINDING, r.LightUBO)
	}
	//might be mvoved to the scene_manager.go
	if r.Root != nil {
		r.Root.Draw(r)
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}
