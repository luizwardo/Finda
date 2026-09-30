package server

import (
	"encoding/json"
	"testing"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/protocol"
)

func TestProcessReportsSearchTime(t *testing.T) {
	h := &Handler{Algo: protocol.AlgoAStar, Map: graph.CityMap()}
	raw, err := json.Marshal(protocol.Request{
		Op:   protocol.OpPathfind,
		From: "NP",
		To:   "AE",
	})
	if err != nil {
		t.Fatal(err)
	}
	resp := h.process(string(raw))
	if !resp.OK {
		t.Fatalf("search failed: %s", resp.Error)
	}
	if resp.ElapsedNs <= 0 {
		t.Fatalf("expected positive search time, got %d", resp.ElapsedNs)
	}
	if resp.Algorithm != protocol.AlgoAStar {
		t.Fatalf("algorithm = %s", resp.Algorithm)
	}
}
