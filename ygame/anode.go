package ygame

import (
	"sort"
	"yam/y3d"
)

type ANode struct {
	Pos        y3d.Vec3
	GlobalGoal float64 //path cost + huristic
	PathCost   float64 //path cost
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
search:
	for len(queue) > 0 && start != end {
		queue.Sort()
		f := queue[0]
		if f.IsVisited {
			queue = queue[1:]
			continue search
		}
		if f.Neighbours != nil && !f.IsObstacle {
		nloop:
			for _, n := range f.Neighbours {
				if n.IsVisited || n.IsObstacle {
					continue nloop
				}
				dist := y3d.Distance(n.Pos, f.Pos) + f.PathCost
				if dist < n.PathCost {
					n.PathCost = dist
					n.GlobalGoal = y3d.Distance(end.Pos, n.Pos) + n.PathCost
					n.Parent = f
					queue = append(queue, n)
				}
			}
		}
		f.IsVisited = true
		//pop the front
		queue = queue[1:]
	}
}
