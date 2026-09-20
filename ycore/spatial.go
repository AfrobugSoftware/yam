package ycore

import (
	"yam/y3d"
)

type SpatialInterface interface {
	UpdateRS(dt float32)
	UpdateGS(dt float32)
	GetParent() SpatialInterface
	SetParent(p SpatialInterface)
	Draw()
	PropagateToRoot()
	GetTransform() *Transform
	GetBoundingBox() y3d.AABB
	GetEffect() Effect
}

type Spatial struct {
	RenderManager    *RenderManager
	Parent           SpatialInterface
	LocalEffect      Effect
	LocalBoundingBox y3d.AABB
	WorldBoundingBox y3d.AABB
	Transform        *Transform
}

func (s *Spatial) GetParent() SpatialInterface {
	return s.Parent
}

func (s *Spatial) SetParent(p SpatialInterface) {
	s.Parent = p
}

func (s *Spatial) GetTransform() *Transform {
	return s.Transform
}

func (s *Spatial) GetBoundingBox() y3d.AABB {
	return s.WorldBoundingBox
}

func (s *Spatial) UpdateRS(dt float32) {

}
func (s *Spatial) UpdateGS(dt float32) {

}

func (s *Spatial) PropagateToRoot() {

}

func (s *Spatial) CalculateScale(factor float32) {
	scaling := (s.LocalBoundingBox.Max.Y - s.LocalBoundingBox.Min.Y) / factor
	s.Transform.SetScale(scaling)
	s.Transform.Recalulate()
}

func (s *Spatial) UpdateWorldBound() {
	s.WorldBoundingBox = s.Transform.TransformAABB(s.LocalBoundingBox)
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
