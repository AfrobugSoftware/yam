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

func (c Capsule) Transform(m Mat4) Capsule {
	vct := Vec3{
		X: m[12],
		Y: m[13],
		Z: m[14],
	}
	m[12], m[13], m[14] = 0.0, 0.0, 0.0
	c.Line.Start = m.MulVec3(c.Line.Start)
	c.Line.End = m.MulVec3(c.Line.End)
	c.Line.Start = Add(c.Line.Start, vct)
	c.Line.End = Add(c.Line.End, vct)

	return c
}
