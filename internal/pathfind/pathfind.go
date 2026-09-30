package pathfind

import (
	"container/heap"
	"math"

	"github.com/wrdo/FInda/internal/graph"
)

// Step records one node settled during search (for client animation).
type Step struct {
	Node string  `json:"node"`
	From string  `json:"from,omitempty"`
	Cost float64 `json:"cost"` // g-score when settled; drives wavefront animation
}

// Result is a completed path search.
type Result struct {
	Path  []string `json:"path"`
	Cost  float64  `json:"cost"`
	Steps []Step   `json:"steps"`
}

type item struct {
	id       string
	priority float64
	index    int
}

type priorityQueue []*item

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].priority < pq[j].priority }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}
func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	it := x.(*item)
	it.index = n
	*pq = append(*pq, it)
}
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	it := old[n-1]
	old[n-1] = nil
	it.index = -1
	*pq = old[:n-1]
	return it
}

// Dijkstra finds the lowest-cost path and records settle order.
func Dijkstra(g *graph.Graph, from, to string) Result {
	return search(g, from, to, false)
}

// AStar finds a path guided by Euclidean heuristic and records settle order.
func AStar(g *graph.Graph, from, to string) Result {
	return search(g, from, to, true)
}

func search(g *graph.Graph, from, to string, useHeuristic bool) Result {
	if _, ok := g.Nodes[from]; !ok {
		return Result{}
	}
	if _, ok := g.Nodes[to]; !ok {
		return Result{}
	}
	if from == to {
		return Result{Path: []string{from}, Cost: 0, Steps: []Step{{Node: from, Cost: 0}}}
	}

	dist := make(map[string]float64, len(g.Nodes))
	prev := make(map[string]string, len(g.Nodes))
	for id := range g.Nodes {
		dist[id] = math.Inf(1)
	}
	dist[from] = 0

	pq := &priorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &item{id: from, priority: 0})

	visited := make(map[string]bool, len(g.Nodes))
	steps := make([]Step, 0, len(g.Nodes))

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*item)
		u := cur.id
		if visited[u] {
			continue
		}
		visited[u] = true
		steps = append(steps, Step{Node: u, From: prev[u], Cost: dist[u]})

		if u == to {
			break
		}

		for _, nb := range g.Neighbors(u) {
			if visited[nb.ID] {
				continue
			}
			alt := dist[u] + nb.Weight
			if alt < dist[nb.ID] {
				dist[nb.ID] = alt
				prev[nb.ID] = u
				prio := alt
				if useHeuristic {
					prio = alt + g.Heuristic(nb.ID, to)
				}
				heap.Push(pq, &item{id: nb.ID, priority: prio})
			}
		}
	}

	if math.IsInf(dist[to], 1) {
		return Result{Steps: steps}
	}

	path := reconstruct(prev, from, to)
	return Result{Path: path, Cost: dist[to], Steps: steps}
}

func reconstruct(prev map[string]string, from, to string) []string {
	var rev []string
	for cur := to; cur != ""; cur = prev[cur] {
		rev = append(rev, cur)
		if cur == from {
			break
		}
	}
	if len(rev) == 0 || rev[len(rev)-1] != from {
		return nil
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}
