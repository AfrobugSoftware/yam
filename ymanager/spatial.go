package ymanager

import (
	"bytes"
	"log"
	"yam/y3d"
)

type SpatialInterface interface {
	UpdateRS(dt float32)
	UpdateGS(dt float32)
	GetParent() SpatialInterface
	GetChildren() []SpatialInterface
	Draw(r *RenderManager)
	PropagateToRoot()
	GetTransform() *Transform
	GetBoundingBox() y3d.AABB
}

type Spatial struct {
	Parent           SpatialInterface
	Children         []SpatialInterface
	LocalBoundingBox y3d.AABB
	WorldBoundingBox y3d.AABB
	Transform        *Transform
	VertexType       string
	DataV, DataI     *bytes.Buffer
	DrawCommand      DrawCommand
	SkinId           int
	StaticBuf        int
}

func NewSpatial(
	parent SpatialInterface,
	boundingBox y3d.AABB,
	tranform *Transform,
	vertexType string,
	dataV, dataI *bytes.Buffer,
	drawCommand DrawCommand,
	skinId int,
	staticBuf int,
) *Spatial {
	return &Spatial{
		Parent:           parent,
		LocalBoundingBox: boundingBox,
		Transform:        tranform,
		VertexType:       vertexType,
		DataV:            dataV,
		DataI:            dataI,
		DrawCommand:      drawCommand,
		SkinId:           skinId,
		StaticBuf:        staticBuf,
	}
}

func (s *Spatial) GetParent() SpatialInterface {
	return s.Parent
}

func (s *Spatial) GetTransform() *Transform {
	return s.Transform
}

func (s *Spatial) GetBoundingBox() y3d.AABB {
	return s.WorldBoundingBox
}

func (s *Spatial) Draw(r *RenderManager) {
	if s.StaticBuf != NO_STATICBUF {
		r.VertextManager.RenderSB(s.StaticBuf, []y3d.Mat4{s.Transform.World})
		return
	}
	if s.DataI != nil && s.DataV != nil {
		err := r.VertextManager.Render(s.VertexType,
			s.DataV, s.DataI, s.SkinId, []y3d.Mat4{s.Transform.World},
			s.DrawCommand)
		if err != nil {
			log.Println(err)
		}
		return
	}
}

func (s *Spatial) UpdateRS(dt float32) {

}
func (s *Spatial) UpdateGS(dt float32) {

}
func (s *Spatial) GetChildren() []SpatialInterface {
	return s.Children
}
func (s *Spatial) PropagateToRoot() {

}

func (s *Spatial) UpdateWorldBound() {
	s.WorldBoundingBox = s.Transform.TransFormAABB(s.LocalBoundingBox)
}
func (s *Spatial) UpdateWorldTransform() {
	if s.Parent != nil {
		s.Transform.UpdateWorld(s.Parent.GetTransform())
	} else {
		s.Transform.World = s.Transform.Local
	}
}
