package protocol

import "github.com/wrdo/FInda/internal/pathfind"

const (
	OpPathfind = "pathfind"
	AlgoDijkstra = "dijkstra"
	AlgoAStar    = "astar"
)

// Request is the JSON payload sent by the client to a server.
type Request struct {
	Op     string `json:"op"`
	From   string `json:"from"`
	To     string `json:"to"`
	Algo   string `json:"algo,omitempty"`
}

// Response is the JSON payload returned by a server.
type Response struct {
	OK        bool             `json:"ok"`
	Error     string           `json:"error,omitempty"`
	Algorithm string           `json:"algorithm"`
	From      string           `json:"from"`
	To        string           `json:"to"`
	Path      []string         `json:"path,omitempty"`
	Cost      float64          `json:"cost,omitempty"`
	Steps     []pathfind.Step  `json:"steps,omitempty"`
}
