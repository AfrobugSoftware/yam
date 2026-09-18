package ycore

import (
	"yam/y3d"
)

type SpatialInterface interface {
	UpdateRS(dt float32)
	UpdateGS(dt float32)
	GetParent() SpatialInterface
	Draw(r *RenderManager)
	PropagateToRoot()
	GetTransform() *Transform
	GetBoundingBox() y3d.OBB
	GetEffect() Effect
}

type Spatial struct {
	Parent           SpatialInterface
	LocalEffect      Effect
	LocalBoundingBox y3d.OBB
	WorldBoundingBox y3d.OBB
	Transform        *Transform
}

func (s *Spatial) GetParent() SpatialInterface {
	return s.Parent
}

func (s *Spatial) GetTransform() *Transform {
	return s.Transform
}

func (s *Spatial) GetBoundingBox() y3d.OBB {
	return s.WorldBoundingBox
}

func (s *Spatial) UpdateRS(dt float32) {

}
func (s *Spatial) UpdateGS(dt float32) {

}

func (s *Spatial) PropagateToRoot() {

}

func (s *Spatial) UpdateWorldBound() {
	s.WorldBoundingBox = s.Transform.TransformOBB(s.LocalBoundingBox)
}
func (s *Spatial) UpdateWorldTransform() {
	if s.Parent != nil {
		s.Transform.UpdateWorld(s.Parent.GetTransform())
	} else {
		s.Transform.World = s.Transform.Local
	}
}

func (s *Spatial) GetEffect() Effect {
	return s.LocalEffect
}
