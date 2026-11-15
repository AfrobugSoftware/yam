package ycore

type PipelineType string
type Pipeline struct {
	Entites  []EntityId
	Program  uint32
	DrawFunc func(r *RenderManager, entites []EntityId)
	Gbuffer  uint32
}
