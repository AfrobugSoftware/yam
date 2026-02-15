package ygame

type Actor interface {
	Update(dt float64)
	Draw()
	ProcessInput()
	GetType() int
}
