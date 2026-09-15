package ycore

type Effect interface {
	Bind(s *ShaderManager)
	Unbind()
}
