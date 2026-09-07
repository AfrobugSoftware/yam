package yecs

import "github.com/go-gl/gl/v4.3-core/gl"

type RenderState struct {
	States []any
}

func (rs *RenderState) AddState(s any) {
	rs.States = append(rs.States, s)
}

type BlendState struct {
	SrcFactor uint32
	DstFactor uint32
	Enable    bool
}

type DepthState struct {
	DepthFunc uint32
	Enable    bool
}

type FaceState struct {
	Enable    bool
	FrontFace uint32
	CullFace  uint32
}

type PolygonMode struct {
	Mode uint32
}

func DisableRenderStates() {
	gl.Disable(gl.BLEND)
	gl.Disable(gl.DEPTH_TEST)
}

func (rs RenderState) SetupRenderState() {
	DisableRenderStates() //reset render states
	for _, r := range rs.States {
		switch v := r.(type) {
		case BlendState:
			if v.Enable {
				gl.Enable(gl.BLEND)
				gl.BlendFunc(v.SrcFactor, v.DstFactor)
			} else {
				gl.Disable(gl.BLEND)
			}
		case DepthState:
			if v.Enable {
				gl.Enable(gl.DEPTH_TEST)
				gl.DepthFunc(v.DepthFunc)
			} else {
				gl.Disable(gl.DEPTH_TEST)
			}
		case FaceState:
			if v.Enable {
				gl.Enable(gl.CULL_FACE)
				gl.CullFace(v.CullFace)
				gl.FrontFace(v.FrontFace)
			} else {
				gl.Disable(gl.CULL_FACE)
			}
		}

	}
}
