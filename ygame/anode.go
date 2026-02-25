package ygame

import (
	"sort"
	"yam/y3d"
)

type ANode struct {
	Pos        y3d.Vec3
	GlobalGoal float64
	LocalGoal  float64
	IsObstacle bool
	IsVisited  bool
	Parent     *ANode
	Neighbours []*ANode
}

type ANodeQueue []*ANode

func (a ANodeQueue) Len() int {
	return len([]*ANode(a))
}

func (a ANodeQueue) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ANodeQueue) Sort() {
	sort.Sort(a)
}

func (a ANodeQueue) Less(i, j int) bool {
	return a[i].GlobalGoal < a[j].GlobalGoal
}

func SolveAStar(graph []ANode, start *ANode, end *ANode) {
	queue := ANodeQueue{}
	start.GlobalGoal = y3d.Distance(end.Pos, start.Pos)
	queue = append(queue, start)
	for len(queue) > 0 {
		queue.Sort()
		f := queue[0]
		if f.Neighbours != nil && !f.IsObstacle {
			for _, r := range f.Neighbours {
				dist := y3d.Distance(r.Pos, f.Pos) + f.LocalGoal
				if dist < r.LocalGoal {
					r.GlobalGoal = y3d.Distance(end.Pos, r.Pos) + dist
					r.LocalGoal = dist
					r.Parent = f
					queue = append(queue, r)
				}
			}
		}
		f.IsVisited = true
		//pop the front
		queue = queue[1:]
	}
}
