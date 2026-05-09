package ygl

import (
	"github.com/go-gl/gl/v4.3-core/gl"
)

type Framebuffer struct {
	Width, Height int32
	fbo           uint32
	Color         []uint32
	Depth         uint32
	DepthTexture  uint32
}

func CreateFrameBuffer(width, height int32, addDepth bool) *Framebuffer {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	f := &Framebuffer{
		fbo:    fbo,
		Width:  width,
		Height: height,
	}
	if addDepth {
		var d uint32
		gl.GenRenderbuffers(1, &d)
		gl.NamedRenderbufferStorage(d, gl.DEPTH_COMPONENT, width, height)
		gl.NamedFramebufferRenderbuffer(fbo, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, d)
		f.Depth = d
	}
	return f
}

func (f *Framebuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, f.fbo)
}

func (f *Framebuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (f *Framebuffer) AttachTexture(texId uint32) {
	attachment := gl.COLOR_ATTACHMENT0 + len(f.Color)
	gl.BindTexture(gl.TEXTURE_2D, texId)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, uint32(attachment), gl.TEXTURE_2D, texId, 0)
	f.Color = append(f.Color, texId)
}

func (f *Framebuffer) AttachDepthTexture(texId uint32) {
	f.DepthTexture = texId
	gl.BindTexture(gl.TEXTURE_2D, texId)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, texId, 0)
}

func (f *Framebuffer) DrawBuffers() {
	drawbuffer := make([]uint32, len(f.Color))
	for i := range f.Color {
		drawbuffer[i] = uint32(gl.COLOR_ATTACHMENT0 + i)
	}
	gl.NamedFramebufferDrawBuffers(f.fbo, int32(len(drawbuffer)), &drawbuffer[0])
}

func (f *Framebuffer) BindFramebufferForWriting() {
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, f.fbo)
	gl.Viewport(0, 0, f.Width, f.Height)
}

func (f *Framebuffer) CheckComplete() bool {
	return gl.CheckFramebufferStatus(gl.FRAMEBUFFER) == gl.FRAMEBUFFER_COMPLETE
}

func DestoryFrameBuffer(f *Framebuffer) {
	gl.DeleteBuffers(1, &f.fbo)
	gl.DeleteTextures(int32(len(f.Color)), &f.Color[0])
	gl.DeleteTextures(1, &f.DepthTexture)
	gl.DeleteRenderbuffers(1, &f.Depth)

	f.fbo = 0
	clear(f.Color)
	f.Depth = 0
	f.DepthTexture = 0
	f.Width = 0
	f.Height = 0
}
