package ycore

import "yam/ywd"

type PipelineType string
type Pipeline struct {
	Entites  []ywd.EntityId
	Program  uint32
	DrawFunc func(r *RenderManager, entites []ywd.EntityId)
	Gbuffer  uint32
}
