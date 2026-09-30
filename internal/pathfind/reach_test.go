package pathfind_test

import (
	"testing"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/pathfind"
)

func TestEveryPlaceIsReachable(t *testing.T) {
	for name, g := range map[string]*graph.Graph{
		"layout":   graph.CityMap(),
		"dijkstra": graph.DijkstraMap(),
		"astar":    graph.AStarMap(),
	} {
		ids := g.NodeIDs()
		if len(ids) != 64 {
			t.Fatalf("%s has %d nodes", name, len(ids))
		}
		for _, a := range ids {
			for _, b := range ids {
				if a == b {
					continue
				}
				r := pathfind.Dijkstra(g, a, b)
				if len(r.Path) == 0 || r.Path[0] != a || r.Path[len(r.Path)-1] != b {
					t.Fatalf("%s: no path %s -> %s (%v)", name, a, b, r.Path)
				}
				aStar := pathfind.AStar(g, a, b)
				if len(aStar.Path) == 0 || aStar.Path[0] != a || aStar.Path[len(aStar.Path)-1] != b {
					t.Fatalf("%s A*: no path %s -> %s (%v)", name, a, b, aStar.Path)
				}
			}
		}
	}
}
