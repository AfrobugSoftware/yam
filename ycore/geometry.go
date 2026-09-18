package ycore

import (
	"bytes"
	"log"
	"yam/y3d"
)

type Geometry struct {
	Spatial
	VertexType       string
	DataV, DataI     *bytes.Buffer
	DrawCommand      DrawCommand
	SkinId           int
	StaticBuf        int
	SkeletalAnimator *SkeletalAnimator
}

func NewGeometry(
	parent SpatialInterface,
	boundingBox y3d.OBB,
	tranform *Transform,
	vertexType string,
	dataV, dataI *bytes.Buffer,
	drawCommand DrawCommand,
	skinId int,
	staticBuf int,
) *Geometry {
	return &Geometry{
		Spatial: Spatial{
			Parent:           parent,
			LocalBoundingBox: boundingBox,
			Transform:        tranform,
		},
		VertexType:  vertexType,
		DataV:       dataV,
		DataI:       dataI,
		DrawCommand: drawCommand,
		SkinId:      skinId,
		StaticBuf:   staticBuf,
	}
}

func (g *Geometry) Draw(r *RenderManager) {
	if g.LocalEffect != nil {
		g.LocalEffect.Bind(r.ShaderManager)
	}
	if g.StaticBuf != NO_STATICBUF {
		r.VertextManager.RenderSB(g.StaticBuf, []y3d.Mat4{g.Transform.World})
		return
	}
	if g.DataI != nil && g.DataV != nil {
		err := r.VertextManager.Render(g.VertexType,
			g.DataV, g.DataI, g.SkinId, []y3d.Mat4{g.Transform.World},
			g.DrawCommand)
		if err != nil {
			log.Println(err)
		}
		return
	}
	if g.LocalEffect != nil {
		g.LocalEffect.Unbind()
	}
}
