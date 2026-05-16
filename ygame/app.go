package ygame

import "time"

type Application interface {
	Startup()
	Draw(currentTime time.Time)
	Shutdown()
}
