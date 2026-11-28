package ycore

import (
	"errors"
	"yam/ywd"

	"github.com/go-gl/gl/v4.3-core/gl"
)

type PipelineType string
type Pipeline struct {
	Name          PipelineType
	Entites       []ywd.EntityId
	Program       uint32
	DrawFunc      func(r *RenderManager, p *Pipeline, entites []ywd.EntityId)
	ScreenProgram uint32
	gBuffer       uint32
}

type GBuffer struct {
	gBuffer     uint32
	gNormal     uint32
	gPosition   uint32
	gAlbedoSpec uint32
	gDepth      uint32
}

func NewPipeline(name PipelineType,
	program uint32,
	gbuffer uint32,
	drawFunc func(r *RenderManager, p *Pipeline, entites []ywd.EntityId)) *Pipeline {
	return &Pipeline{
		Name:     name,
		Program:  program,
		gBuffer:  gbuffer,
		DrawFunc: drawFunc,
	}
}

func (g *GBuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, g.gBuffer)
}

func (g *GBuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (p *GBuffer) CreateGbuffer(width, height int32) error {
	gl.GenFramebuffers(1, &p.gBuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.gBuffer)

	//position
	gl.CreateTextures(gl.TEXTURE_2D, 1, &p.gPosition)
	gl.TextureStorage2D(p.gPosition, 1, gl.RGBA16F, width, height)
	gl.TextureParameteri(p.gPosition, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(p.gPosition, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p.gPosition, 0)

	//normals
	gl.CreateTextures(gl.TEXTURE_2D, 1, &p.gNormal)
	gl.TextureStorage2D(p.gNormal, 1, gl.RGBA16F, width, height)
	gl.TextureParameteri(p.gNormal, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(p.gNormal, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, p.gNormal, 0)

	//alebo spec
	gl.CreateTextures(gl.TEXTURE_2D, 1, &p.gAlbedoSpec)
	gl.TextureStorage2D(p.gAlbedoSpec, 1, gl.RGBA16F, width, height)
	gl.TextureParameteri(p.gAlbedoSpec, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(p.gAlbedoSpec, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, p.gAlbedoSpec, 0)

	//depth texture
	gl.CreateTextures(gl.TEXTURE_2D, 1, &p.gDepth)
	gl.TextureStorage2D(p.gDepth, 1, gl.DEPTH24_STENCIL8, width, height)
	gl.TextureParameteri(p.gDepth, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TextureParameteri(p.gDepth, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.TEXTURE_2D, p.gDepth, 0)

	attachments := [3]uint32{
		gl.COLOR_ATTACHMENT0,
		gl.COLOR_ATTACHMENT1,
		gl.COLOR_ATTACHMENT2,
	}
	gl.DrawBuffers(3, &attachments[0])
	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		return errors.New("failed to create g-buffer")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	return nil
}
