package main

import (
	"testing"

	"fyne.io/fyne/v2"

	"github.com/wrdo/FInda/internal/graph"
)

func TestHitPlacePicksCircleAndRoad(t *testing.T) {
	g := graph.CityMap()
	size := fyne.NewSize(900, 580)

	if got := hitPlace(g, size, fyne.NewPos(120, 70)); got != "NP" {
		t.Fatalf("circle hit = %s", got)
	}

	mid := fyne.NewPos(220, 65)
	hit := hitTarget(g, size, mid)
	if !hit.OnEdge {
		t.Fatalf("expected an edge hit near the road midpoint, got %#v", hit)
	}
	if hit.Place != "NP" && hit.Place != "BI" {
		t.Fatalf("road place = %s, want NP or BI", hit.Place)
	}
	if (hit.EdgeA != "NP" || hit.EdgeB != "BI") && (hit.EdgeA != "BI" || hit.EdgeB != "NP") {
		t.Fatalf("edge = %s-%s", hit.EdgeA, hit.EdgeB)
	}
}
