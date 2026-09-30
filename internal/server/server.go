package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/pathfind"
	"github.com/wrdo/FInda/internal/protocol"
)

// Handler serves pathfinding over TCP (one goroutine per connection).
type Handler struct {
	Algo string
	Map  *graph.Graph
}

// ListenAndServe binds addr and accepts connections concurrently.
func (h *Handler) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()
	log.Printf("[%s] listening on %s", h.Algo, addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[%s] accept error: %v", h.Algo, err)
			continue
		}
		go h.handle(conn)
	}
}

func (h *Handler) handle(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	log.Printf("[%s] connection from %s", h.Algo, remote)

	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		resp := h.process(line)
		enc, err := json.Marshal(resp)
		if err != nil {
			fmt.Fprintln(conn, `{"ok":false,"error":"encode failed"}`)
			continue
		}
		if _, err := fmt.Fprintln(conn, string(enc)); err != nil {
			log.Printf("[%s] write error to %s: %v", h.Algo, remote, err)
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("[%s] read error from %s: %v", h.Algo, remote, err)
	}
	log.Printf("[%s] closed %s", h.Algo, remote)
}

func (h *Handler) process(line string) protocol.Response {
	var req protocol.Request
	if err := json.Unmarshal([]byte(line), &req); err != nil {
		return protocol.Response{OK: false, Error: "invalid JSON", Algorithm: h.Algo}
	}
	if req.Op != protocol.OpPathfind {
		return protocol.Response{OK: false, Error: "unsupported op", Algorithm: h.Algo}
	}
	if req.From == "" || req.To == "" {
		return protocol.Response{OK: false, Error: "from/to required", Algorithm: h.Algo}
	}
	if _, ok := h.Map.Nodes[req.From]; !ok {
		return protocol.Response{OK: false, Error: "unknown place: " + req.From, Algorithm: h.Algo, From: req.From, To: req.To}
	}
	if _, ok := h.Map.Nodes[req.To]; !ok {
		return protocol.Response{OK: false, Error: "unknown place: " + req.To, Algorithm: h.Algo, From: req.From, To: req.To}
	}

	start := time.Now()
	var result pathfind.Result
	switch h.Algo {
	case protocol.AlgoAStar:
		result = pathfind.AStar(h.Map, req.From, req.To)
	default:
		result = pathfind.Dijkstra(h.Map, req.From, req.To)
	}
	elapsed := time.Since(start).Nanoseconds()

	if len(result.Path) == 0 {
		return protocol.Response{
			OK:        false,
			Error:     "no path found",
			Algorithm: h.Algo,
			From:      req.From,
			To:        req.To,
			Steps:     result.Steps,
			ElapsedNs: elapsed,
		}
	}

	return protocol.Response{
		OK:        true,
		Algorithm: h.Algo,
		From:      req.From,
		To:        req.To,
		Path:      result.Path,
		Cost:      result.Cost,
		Steps:     result.Steps,
		ElapsedNs: elapsed,
	}
}
