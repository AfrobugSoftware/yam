package ygame

import "time"

type Application interface {
	Startup()
	Update(currentTime time.Time)
	Draw()
	Shutdown()
}
