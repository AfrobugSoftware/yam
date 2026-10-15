package ycore

import (
	"bytes"
	"encoding/gob"
	"log"
	"yam/y3d"
	"yam/ycontroller"
)

type Geometry struct {
	Spatial
	VertexType       string
	DrawCommand      DrawCommand
	InstanceCount    int
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
	skinId int,
	staticBuf int,
	movec *ycontroller.MovementController,
) *Geometry {
	g := &Geometry{
		Spatial: Spatial{
			RenderManager:    renderManager,
			Parent:           parent,
			LocalBoundingBox: boundingBox,
			Transform:        tranform,
			MoveController:   movec,
		},
		VertexType:    vertexType,
		InstanceCount: 1,
		SkinId:        skinId,
		StaticBuf:     staticBuf,
	}
	g.Setup(dataV, dataI)
	return g
}
func (g *Geometry) Setup(dataV, dataI *bytes.Buffer) {
	if dataI != nil && dataV != nil && g.StaticBuf == NO_STATICBUF {
		err := g.RenderManager.VertextManager.LoadCache(g.VertexType,
			dataV, dataI, g.SkinId,
			&g.DrawCommand)
		if err != nil {
			log.Println(err)
		}
		return
	}
}

func (g *Geometry) Draw() {
	if g.SkeletalAnimator != nil {
		g.SkeletalAnimator.BindBuffer()
	}
	if g.StaticBuf != NO_STATICBUF {
		if g.LocalEffect != nil {
			g.LocalEffect.Bind(g.RenderManager.ShaderManager)
		}
		g.RenderManager.VertextManager.RenderSB(g.StaticBuf, g.Transform.World)
		if g.LocalEffect != nil {
			g.LocalEffect.Unbind()
		}
		return
	} else {
		g.RenderManager.VertextManager.LoadInstances(g.VertexType, g.SkinId, &g.DrawCommand, g.InstanceCount)
		g.RenderManager.VertextManager.LoadDrawCommand(g.VertexType, g.SkinId, g.DrawCommand)
		g.RenderManager.VertextManager.LoadMatrix(
			g.VertexType,
			int(g.DrawCommand.BaseInstance),
			g.SkinId,
			g.Transform.World,
		)
	}
}
