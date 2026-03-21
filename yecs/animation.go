package yecs

type Animation2D struct {
	FPS          float64
	CurrentFrame int
	Ticks        int64
	TotalFrames  int
}

func (a *Animation2D) GetNextFrame(dt float64) int {
	into := a.FPS * dt
	a.CurrentFrame += int(into)
	return a.CurrentFrame % a.TotalFrames
}
