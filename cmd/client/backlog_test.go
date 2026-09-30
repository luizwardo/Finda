package main

import (
	"strings"
	"testing"
	"time"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/protocol"
)

func TestFormatBacklogShowsServerCostsThenTotal(t *testing.T) {
	g := graph.CityMap()
	m := NewMapView(g)
	text := formatBacklog(
		time.Date(2026, 9, 30, 15, 4, 5, 0, time.Local),
		m, "AE", "UN",
		protocol.Response{OK: true, Cost: 1234.5},
		protocol.Response{OK: true, Cost: 1102},
		nil, nil,
	)
	want := "Dijkstra 1234.5 · A* 1102.0 · Total 2336.5"
	if text != want {
		t.Fatalf("got %q want %q", text, want)
	}
	if i, j := strings.Index(text, "Dijkstra"), strings.Index(text, "Total"); i < 0 || j < 0 || i > j {
		t.Fatalf("server costs should appear before total: %s", text)
	}
}
