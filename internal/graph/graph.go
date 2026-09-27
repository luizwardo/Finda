package graph

import (
	"math"
	"sort"
)

// Node is a named location on the city map with canvas coordinates.
type Node struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	X    float32 `json:"x"`
	Y    float32 `json:"y"`
}

// Edge is a bidirectional road between two nodes.
type Edge struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Weight float64 `json:"weight"`
}

// Graph holds the city topology used by client and servers.
type Graph struct {
	Nodes map[string]Node
	Adj   map[string][]neighbor
}

type neighbor struct {
	ID     string
	Weight float64
}

// New builds an empty graph.
func New() *Graph {
	return &Graph{
		Nodes: make(map[string]Node),
		Adj:   make(map[string][]neighbor),
	}
}

// AddNode registers a location.
func (g *Graph) AddNode(n Node) {
	g.Nodes[n.ID] = n
}

// AddEdge adds a bidirectional road. Weight defaults to Euclidean distance when <= 0.
func (g *Graph) AddEdge(from, to string, weight float64) {
	if weight <= 0 {
		a, b := g.Nodes[from], g.Nodes[to]
		weight = Euclidean(a.X, a.Y, b.X, b.Y)
	}
	g.Adj[from] = append(g.Adj[from], neighbor{ID: to, Weight: weight})
	g.Adj[to] = append(g.Adj[to], neighbor{ID: from, Weight: weight})
}

// Neighbors returns adjacent node IDs and edge weights.
func (g *Graph) Neighbors(id string) []neighbor {
	return g.Adj[id]
}

// NodeIDs returns a sorted list of IDs for UI selects.
func (g *Graph) NodeIDs() []string {
	ids := make([]string, 0, len(g.Nodes))
	for id := range g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// DisplayNames returns "Name (ID)" labels aligned with NodeIDs order.
func (g *Graph) DisplayNames() []string {
	ids := g.NodeIDs()
	out := make([]string, len(ids))
	for i, id := range ids {
		n := g.Nodes[id]
		out[i] = n.Name + " (" + id + ")"
	}
	return out
}

// Euclidean distance between two points.
func Euclidean(x1, y1, x2, y2 float32) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Hypot(dx, dy)
}

// Heuristic returns Euclidean distance between two node IDs (for A*).
func (g *Graph) Heuristic(a, b string) float64 {
	na, okA := g.Nodes[a]
	nb, okB := g.Nodes[b]
	if !okA || !okB {
		return 0
	}
	return Euclidean(na.X, na.Y, nb.X, nb.Y)
}
