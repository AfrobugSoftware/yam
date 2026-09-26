package ygame

type Application interface {
	Startup(e *Engine)
	Update(currentTime float64)
	Draw()
	Shutdown()
}
