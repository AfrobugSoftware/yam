package ygame

import (
	"math"
	"slices"
	"sort"
	"yam/y3d"
)

type ANode struct {
	Pos        y3d.Vec3
	GlobalGoal float64 //path cost + huristic
	PathCost   float64 //path cost
	IsObstacle bool
	IsVisited  bool
	Parent     int
	Indx       int
	Neighbours []int
}

func NewANode(pos y3d.Vec3, isObstacle bool, indx int, neighbours []int) ANode {
	return ANode{
		Pos:        pos,
		GlobalGoal: math.MaxFloat64,
		PathCost:   math.MaxFloat64,
		IsObstacle: isObstacle,
		IsVisited:  false,
		Indx:       indx,
		Neighbours: neighbours,
	}
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

func SolveAStar(graph []ANode, start int, end int) []y3d.Vec3 {
	queue := ANodeQueue{}
	graph[start].GlobalGoal = y3d.Distance(graph[end].Pos, graph[start].Pos)
	queue = append(queue, &graph[start])
	cur := start
search:
	for len(queue) > 0 && cur != end {
		queue.Sort()
		f := queue[0]
		cur = f.Indx
		if f.IsVisited {
			queue = queue[1:]
			continue search
		}
		if f.Neighbours != nil && !f.IsObstacle {
		nloop:
			for _, n := range f.Neighbours {
				if graph[n].IsVisited || graph[n].IsObstacle {
					continue nloop
				}
				dist := y3d.Distance(graph[n].Pos, f.Pos) + f.PathCost
				if dist < graph[n].PathCost {
					graph[n].PathCost = dist
					graph[n].GlobalGoal = y3d.Distance(graph[end].Pos, graph[n].Pos) + graph[n].PathCost
					graph[n].Parent = f.Indx
					queue = append(queue, &graph[n])
				}
			}
		}
		f.IsVisited = true
		//pop the front
		queue = queue[1:]
	}
	if cur == end {
		path := make([]y3d.Vec3, 0, len(graph))
		i := end
		for i != start {
			n := graph[i]
			path = append(path, n.Pos)
			i = n.Parent
		}
		slices.Reverse(path)
		return path
	}
	return nil
}
