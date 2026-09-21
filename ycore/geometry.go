package ycore

import (
	"bytes"
	"encoding/gob"
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

func (g *Geometry) Write(e *gob.Encoder) error {
	return e.Encode(*g)
}

func (g *Geometry) Read(d *gob.Decoder) error {
	return d.Decode(g)
}

func NewGeometry(
	renderManager *RenderManager,
	parent SpatialInterface,
	boundingBox y3d.AABB,
	tranform *Transform,
	vertexType string,
	dataV, dataI *bytes.Buffer,
	drawCommand DrawCommand,
	skinId int,
	staticBuf int,
) *Geometry {
	g := &Geometry{
		Spatial: Spatial{
			RenderManager:    renderManager,
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
	g.Setup()
	return g
}
func (g *Geometry) Setup() {
	if g.DataI != nil && g.DataV != nil {
		err := g.RenderManager.VertextManager.LoadCache(g.VertexType,
			g.DataV, g.DataI, g.SkinId,
			&g.DrawCommand)
		if err != nil {
			log.Println(err)
		}
		return
	}
}

func (g *Geometry) Draw() {
	if g.LocalEffect != nil {
		g.LocalEffect.Bind(g.RenderManager.ShaderManager)
	}
	// if g.StaticBuf != NO_STATICBUF {
	// 	g.RenderManager.VertextManager.RenderSB(g.StaticBuf, []y3d.Mat4{g.Transform.World})
	// 	return
	// } else {
	// }
	g.RenderManager.VertextManager.LoadMatrix(
		g.VertexType, int(g.DrawCommand.BaseInstance),
		g.SkinId, []y3d.Mat4{g.Transform.World},
	)
	if g.LocalEffect != nil {
		g.LocalEffect.Unbind()
	}
}
