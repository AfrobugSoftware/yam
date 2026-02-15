package y3d

type AABB struct {
	MinX, MinY float64
	MaxX, MaxY float64
}

func (a AABB) Overlaps(b AABB) bool {
	return a.MinX <= b.MaxX &&
		a.MaxX >= b.MinX &&
		a.MinY <= b.MaxY &&
		a.MaxY >= b.MinY
}
