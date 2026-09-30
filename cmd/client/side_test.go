package main

import "testing"

func TestSideAmountKeepsCanonicalSide(t *testing.T) {
	// edgeKey sorts to AE-PR. Travel AE→PR keeps +side; PR→AE flips.
	along := &travel{from: "AE", node: "PR"}
	against := &travel{from: "PR", node: "AE"}
	if sideAmount(along, 4.5) != 4.5 {
		t.Fatalf("along = %v", sideAmount(along, 4.5))
	}
	if sideAmount(against, 4.5) != -4.5 {
		t.Fatalf("against = %v", sideAmount(against, 4.5))
	}
	if sideAmount(along, -4.5) != -4.5 {
		t.Fatalf("astar along = %v", sideAmount(along, -4.5))
	}
	if sideAmount(against, -4.5) != 4.5 {
		t.Fatalf("astar against = %v", sideAmount(against, -4.5))
	}
}
