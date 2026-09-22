package ygame

type Application interface {
	Startup()
	Update(currentTime float64)
	Draw()
	Shutdown()
}
