package y3d

type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

func (a Rect) Overlaps(b Rect) bool {
	no := (a.Width+a.X < b.X ||
		a.Height+a.Y < b.Y ||
		b.Width+b.X < a.X ||
		b.Height+b.Y < a.Y)
	return !no
}
