package y3d

type Polygon struct {
	Plane       Plane
	BoundingBox AABB
	Flag        uint32
	Points      []Vec3
	Indices     []uint
}

func (p Polygon) CalculateBoundingBox() {
	min := p.Points[0]
	max := min
	for i := range len(p.Points) {
		if p.Points[i].X > max.X {
			max.X = p.Points[i].X
		} else if p.Points[i].X < min.X {
			min.X = p.Points[i].X
		}

		if p.Points[i].Y > max.Y {
			max.Y = p.Points[i].Y
		} else if p.Points[i].Y < min.Y {
			min.Y = p.Points[i].Y
		}

		if p.Points[i].Z > max.Z {
			max.Z = p.Points[i].Z
		} else if p.Points[i].Z < min.Z {
			min.Z = p.Points[i].Z
		}
	}
	p.BoundingBox.Max = max
	p.BoundingBox.Min = min
}

func NewPolygon(points []Vec3, indices []uint) Polygon {
	if len(points) == 0 || len(indices) == 0 {
		panic("cannot create a polygon without data")
	}
	var p Polygon
	copy(p.Points, points)
	copy(p.Indices, indices)

	gotThem := false
	var edge [2]Vec3
	edge[0] = Sub(points[1], points[0])
	for i := 2; !gotThem; i++ {
		if i+1 >= len(points) {
			break
		}
		edge[1] = Sub(points[i], points[0])
		edge[0] = Normalize(edge[0])
		edge[1] = Normalize(edge[1])

		if Dot(edge[0], edge[1]) != 0 {
			gotThem = true
		}
	}
	if !gotThem {
		panic("ill formed polygon - all points a co-linear")
	}
	p.Plane.N = Normalize(Cross(edge[1], edge[0]))
	p.Plane.D = -Dot(p.Plane.N, points[0])
	p.CalculateBoundingBox()
	return p
}

func (p *Polygon) SwapFaces() {
	for i, j := 0, len(p.Indices)-1; i < j; i, j = i+1, j-1 {
		p.Indices[i], p.Indices[j] = p.Indices[j], p.Indices[i]
	}
	p.Plane.N = Smul(p.Plane.N, -1.0)
	p.Plane.D *= -1.0
}

func (p Polygon) ClipAABB(aabb AABB) Polygon {
	planes := aabb.GetPlanes()
	for _, pl := range planes {
		if pl.ClassifyPolygon(p) == CLIPPED {
			_, back := p.Clip(pl)
			p = back
		}
	}
	return p
}

// always classify the polygon on a plain before calling this method
func (p Polygon) Clip(plane Plane) (front Polygon, back Polygon) {
	pFront := make([]Vec3, 0)
	pBack := make([]Vec3, 0)
	//classify first point
	switch plane.Classify(p.Points[0]) {
	case FRONT:
		pFront = append(pFront, p.Points[0])
	case BACK:
		pBack = append(pBack, p.Points[0])
	case PLANER:
		pFront = append(pFront, p.Points[0])
		pBack = append(pBack, p.Points[0])
	default:
		panic("ill formed polygon")
	}

	for loop := 1; loop < len(p.Points)+1; loop++ {
		var current int
		if loop == len(p.Points) {
			current = 0
		} else {
			current = loop
		}
		a := p.Points[loop-1]
		b := p.Points[current]
		classA := p.Plane.Classify(a)
		classB := p.Plane.Classify(b)
		if classB == PLANER {
			pBack = append(pBack, p.Points[current])
			pFront = append(pFront, p.Points[current])
		} else {
			line := LineSegment{
				Start: a,
				End:   b,
			}
			t, hit := line.IntersectsPlane(p.Plane)
			if hit && classA != PLANER {
				hitPoint := line.PointOnLine(t)
				pBack = append(pBack, hitPoint)
				pFront = append(pFront, hitPoint)
			}
			if current == 0 {
				continue
			}
			switch classB {
			case FRONT:
				pFront = append(pFront, p.Points[current])
			case BACK:
				pBack = append(pBack, p.Points[current])
			}
		}
	}
	var i0, i1, i2 uint
	if len(pFront) > 2 {
		pnFront := make([]uint, 0, (len(pFront)-2)*3)
		for loop := 0; loop < len(pFront)-2; loop++ {
			if loop == 0 {
				i0 = 0
				i1 = 1
				i2 = 2
			} else {
				i1 = i2
				i2++
			}
			pnFront[(loop * 3)] = i0
			pnFront[(loop*3)+1] = i1
			pnFront[(loop*3)+2] = i2
		}
		front = NewPolygon(pFront, pnFront)
		if Dot(front.Plane.N, p.Plane.N) < 0 {
			(&front).SwapFaces()
		}
	}
	if len(pBack) > 2 {
		pnBack := make([]uint, 0, (len(pBack)-2)*3)
		for loop := 0; loop < len(pBack)-2; loop++ {
			if loop == 0 {
				i0 = 0
				i1 = 1
				i2 = 2
			} else {
				i1 = i2
				i2++
			}
			pnBack[(loop * 3)] = i0
			pnBack[(loop*3)+1] = i1
			pnBack[(loop*3)+2] = i2
		}
		back = NewPolygon(pBack, pnBack)
		if Dot(back.Plane.N, p.Plane.N) < 0 {
			(&back).SwapFaces()
		}
	}
	return front, back
}
