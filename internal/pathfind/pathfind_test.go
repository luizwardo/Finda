package pathfind_test

import (
	"testing"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/pathfind"
)

func TestDijkstraAndAStarReachAirport(t *testing.T) {
	g := graph.CityMap()
	d := pathfind.Dijkstra(g, "NP", "AE")
	a := pathfind.AStar(g, "NP", "AE")
	if len(d.Path) == 0 {
		t.Fatal("dijkstra: empty path")
	}
	if len(a.Path) == 0 {
		t.Fatal("astar: empty path")
	}
	if d.Path[0] != "NP" || d.Path[len(d.Path)-1] != "AE" {
		t.Fatalf("dijkstra path ends wrong: %v", d.Path)
	}
	if a.Path[0] != "NP" || a.Path[len(a.Path)-1] != "AE" {
		t.Fatalf("astar path ends wrong: %v", a.Path)
	}
	if d.Cost <= 0 || a.Cost <= 0 {
		t.Fatalf("invalid costs d=%v a=%v", d.Cost, a.Cost)
	}
	if len(d.Steps) == 0 || len(a.Steps) == 0 {
		t.Fatal("expected exploration steps for animation")
	}
	// A* should typically explore fewer or equal nodes on this map
	t.Logf("dijkstra cost=%.1f steps=%d path=%v", d.Cost, len(d.Steps), d.Path)
	t.Logf("astar    cost=%.1f steps=%d path=%v", a.Cost, len(a.Steps), a.Path)
}

func TestServerMapsDisagreeOnAirportToUniversity(t *testing.T) {
	dMap := graph.DijkstraMap()
	aMap := graph.AStarMap()

	dOnD := pathfind.Dijkstra(dMap, "AE", "UN")
	aOnD := pathfind.AStar(dMap, "AE", "UN")
	dOnA := pathfind.Dijkstra(aMap, "AE", "UN")
	aOnA := pathfind.AStar(aMap, "AE", "UN")

	if dOnD.Cost != aOnD.Cost {
		t.Fatalf("Dijkstra map: algorithms disagree, dijkstra %.1f astar %.1f", dOnD.Cost, aOnD.Cost)
	}
	if dOnA.Cost != aOnA.Cost {
		t.Fatalf("A* map: algorithms disagree, dijkstra %.1f astar %.1f", dOnA.Cost, aOnA.Cost)
	}
	if dOnD.Cost == dOnA.Cost {
		t.Fatalf("both maps returned the same cost %.1f", dOnD.Cost)
	}
	t.Logf("dijkstra map cost=%.0f path=%v", dOnD.Cost, dOnD.Path)
	t.Logf("astar map    cost=%.0f path=%v", dOnA.Cost, dOnA.Path)
}
