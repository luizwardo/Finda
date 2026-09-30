package graph

import (
	"math"
	"testing"
)

func TestServerMapsUseDifferentDistances(t *testing.T) {
	dijk := DijkstraMap()
	star := AStarMap()
	layout := CityMap()

	if len(layout.Nodes) != 64 || len(dijk.Nodes) != 64 || len(star.Nodes) != 64 {
		t.Fatalf("expected 64 places, got layout %d dijkstra %d astar %d", len(layout.Nodes), len(dijk.Nodes), len(star.Nodes))
	}
	for _, id := range []string{"NP", "BI", "HO", "TE", "PC", "SH", "ES", "MU", "UN", "ME", "PR", "BR", "ST", "AR", "AE", "LI"} {
		if _, ok := layout.Nodes[id]; !ok {
			t.Fatalf("missing original place %s", id)
		}
	}

	if len(dijkstraWeights) != len(astarWeights) {
		t.Fatalf("weight tables differ in size: %d vs %d", len(dijkstraWeights), len(astarWeights))
	}
	for key, w := range dijkstraWeights {
		other, ok := astarWeights[key]
		if !ok {
			t.Fatalf("A* map is missing %s", key)
		}
		if w == other {
			t.Fatalf("road %s has the same distance on both maps: %v", key, w)
		}
		a, b := splitKey(key)
		floor := Euclidean(layout.Nodes[a].X, layout.Nodes[a].Y, layout.Nodes[b].X, layout.Nodes[b].Y)
		if w < floor || other < floor {
			t.Fatalf("road %s undercuts the straight line (%.1f): dijkstra %v, astar %v", key, floor, w, other)
		}
	}
	for id, nbs := range dijk.Adj {
		for _, nb := range nbs {
			dw := nb.Weight
			var aw float64
			var found bool
			for _, other := range star.Neighbors(id) {
				if other.ID == nb.ID {
					aw = other.Weight
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("A* map is missing %s-%s", id, nb.ID)
			}
			if dw == aw {
				t.Fatalf("road %s-%s has the same distance on both maps: %v", id, nb.ID, dw)
			}
			floor := Euclidean(layout.Nodes[id].X, layout.Nodes[id].Y, layout.Nodes[nb.ID].X, layout.Nodes[nb.ID].Y)
			if dw < floor || aw < floor {
				t.Fatalf("road %s-%s undercuts the straight line (%.1f): dijkstra %v, astar %v", id, nb.ID, floor, dw, aw)
			}
		}
	}

	viaPort := []string{"AE", "PR", "ME", "UN"}
	viaArena := []string{"AE", "AR", "ME", "UN"}
	for _, g := range []*Graph{dijk, star} {
		port, arena := pathCost(t, g, viaPort), pathCost(t, g, viaArena)
		if port == arena {
			t.Fatalf("porto and arena routes cost the same (%v)", port)
		}
	}
	if pathCost(t, dijk, viaPort) == pathCost(t, star, viaPort) {
		t.Fatal("porto route has the same cost on both maps")
	}
	if pathCost(t, dijk, viaArena) == pathCost(t, star, viaArena) {
		t.Fatal("arena route has the same cost on both maps")
	}
}

func splitKey(key string) (string, string) {
	for i := 0; i < len(key); i++ {
		if key[i] == '-' {
			return key[:i], key[i+1:]
		}
	}
	return key, ""
}

func pathCost(t *testing.T, g *Graph, path []string) float64 {
	t.Helper()
	var sum float64
	for i := 1; i < len(path); i++ {
		var found bool
		for _, nb := range g.Neighbors(path[i-1]) {
			if nb.ID == path[i] {
				sum += nb.Weight
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing road %s-%s", path[i-1], path[i])
		}
	}
	if math.IsNaN(sum) || sum <= 0 {
		t.Fatalf("bad path cost %v", sum)
	}
	return sum
}
