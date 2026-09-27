package main

import (
	"flag"
	"log"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/protocol"
	"github.com/wrdo/FInda/internal/server"
)

func main() {
	algo := flag.String("algo", protocol.AlgoDijkstra, "algorithm: dijkstra | astar")
	addr := flag.String("addr", ":9001", "listen address, e.g. :9001")
	flag.Parse()

	switch *algo {
	case protocol.AlgoDijkstra, protocol.AlgoAStar:
	default:
		log.Fatalf("unknown algo %q (use dijkstra or astar)", *algo)
	}

	h := &server.Handler{
		Algo: *algo,
		Map:  graph.CityMap(),
	}
	log.Fatal(h.ListenAndServe(*addr))
}
