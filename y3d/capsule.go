package y3d

type Capsule struct {
	Line   LineSegment
	Radius float32
}

func (c Capsule) Contains(point Vec3) bool {
	distSq := c.Line.MinDistSq(point)
	return distSq <= (c.Radius * c.Radius)
}

func (c Capsule) Intersects(d Capsule) bool {
	distSq := c.Line.MinLineSegmentDistSq(d.Line)
	sumRadii := c.Radius + d.Radius
	return distSq <= (sumRadii * sumRadii)
}
