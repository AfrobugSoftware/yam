package yecs

import gl "github.com/chsc/gogl/gl33"

type RenderState struct {
	States []any
}

func (rs *RenderState) AddState(s any) {
	rs.States = append(rs.States, s)
}

type BlendState struct {
	SrcFactor gl.Enum
	DstFactor gl.Enum
	Enable    bool
}

type DepthState struct {
	DepthFunc gl.Enum
	Enable    bool
}

func (rs RenderState) SetupRenderState() {
	for _, r := range rs.States {
		switch v := r.(type) {
		case BlendState:
			if v.Enable {
				gl.Enable(gl.BLEND)
			} else {
				gl.Disable(gl.BLEND)
			}
			gl.BlendFunc(v.SrcFactor, v.DstFactor)
		case DepthState:
			if v.Enable {
				gl.Enable(gl.DEPTH_TEST)
			} else {
				gl.Disable(gl.DEPTH_TEST)
			}
			gl.DepthFunc(v.DepthFunc)
		}
	}
}
