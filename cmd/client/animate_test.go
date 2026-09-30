package main

import (
	"testing"

	"github.com/wrdo/FInda/internal/pathfind"
)

func TestWaveAtSpreadsMultipleRoads(t *testing.T) {
	steps := []pathfind.Step{
		{Node: "A", Cost: 0},
		{Node: "B", From: "A", Cost: 10},
		{Node: "C", From: "A", Cost: 12},
		{Node: "D", From: "B", Cost: 20},
	}
	origin, done, active := waveAt(nil, steps, 11)
	if origin != "A" {
		t.Fatalf("origin = %s", origin)
	}
	if len(done) != 2 { // A and B
		t.Fatalf("done = %d %#v", len(done), done)
	}
	if len(active) != 2 {
		t.Fatalf("expected C and D in flight, got %#v", active)
	}

	_, done, active = waveAt(nil, steps, 6)
	if len(done) != 1 || done[0].Node != "A" {
		t.Fatalf("early done = %#v", done)
	}
	if len(active) != 2 {
		t.Fatalf("expected B and C in flight, got %#v", active)
	}
	for _, a := range active {
		if a.step.Node != "B" && a.step.Node != "C" {
			t.Fatalf("unexpected active %#v", a)
		}
	}
}
