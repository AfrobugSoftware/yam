package ymanager

import "github.com/go-gl/gl/v4.3-core/gl"

type RenderState interface {
	Activate()
	Deactivate()
}

type BlendState struct {
	SrcFactor uint32
	DstFactor uint32
	Enable    bool
}

func (b BlendState) Activate() {
	if b.Enable {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(b.SrcFactor, b.DstFactor)
	}
}

func (b BlendState) Deactivate() {
	if b.Enable {
		gl.Disable(gl.BLEND)
	}
}

type DepthState struct {
	DepthFunc uint32
	Enable    bool
}

func (d DepthState) Activate() {
	if d.Enable {
		gl.Enable(gl.DEPTH_TEST)
		gl.DepthFunc(d.DepthFunc)
	}
}

func (d DepthState) Deactivate() {
	if d.Enable {
		gl.Disable(gl.DEPTH_TEST)
	}
}

type FaceState struct {
	Enable    bool
	FrontFace uint32
	CullFace  uint32
}

func (f FaceState) Activate() {
	if f.Enable {
		gl.Enable(gl.CULL_FACE)
		gl.CullFace(f.CullFace)
	}
}

func (f FaceState) Deactivate() {
	if f.Enable {
		gl.Disable(gl.CULL_FACE)
	}
}

type PolygonMode struct {
	Mode uint32
}

func (p PolygonMode) Activate() {
	gl.PolygonMode(gl.FRONT_AND_BACK, p.Mode)
}
func (p PolygonMode) Deactivate() {
	gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
}
