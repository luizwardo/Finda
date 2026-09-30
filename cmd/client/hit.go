package main

import (
	"math"

	"fyne.io/fyne/v2"

	"github.com/wrdo/FInda/internal/graph"
)

func mapTransform(size fyne.Size) (scale, ox, oy float32, ok bool) {
	if size.Width < 1 || size.Height < 1 {
		return 0, 0, 0, false
	}
	sx := size.Width / 900
	sy := size.Height / 580
	scale = float32(math.Min(float64(sx), float64(sy)))
	ox = (size.Width - 900*scale) / 2
	oy = (size.Height - 580*scale) / 2
	return scale, ox, oy, true
}

// mapHit is the place (and optional road) under a click.
type mapHit struct {
	Place  string
	EdgeA  string
	EdgeB  string
	OnEdge bool
}

// hitTarget prefers a road when the pointer is close to one, otherwise a place.
func hitTarget(g *graph.Graph, size fyne.Size, at fyne.Position) mapHit {
	if g == nil {
		return mapHit{}
	}
	scale, ox, oy, ok := mapTransform(size)
	if !ok {
		return mapHit{}
	}
	pos := func(n graph.Node) fyne.Position {
		return fyne.NewPos(ox+n.X*scale, oy+n.Y*scale)
	}

	edgeReach := 14 * scale
	bestEdge := edgeReach
	var edgeA, edgeB string
	var edgeT float32
	seen := make(map[[2]string]bool)
	for from, nbs := range g.Adj {
		for _, nb := range nbs {
			key := edgeKey(from, nb.ID)
			if seen[key] {
				continue
			}
			seen[key] = true
			a, okA := g.Nodes[from]
			b, okB := g.Nodes[nb.ID]
			if !okA || !okB {
				continue
			}
			d, t := distToSegment(at, pos(a), pos(b))
			if d > bestEdge {
				continue
			}
			bestEdge = d
			edgeA, edgeB, edgeT = from, nb.ID, t
		}
	}
	if edgeA != "" {
		place := edgeA
		if edgeT >= 0.5 {
			place = edgeB
		}
		return mapHit{Place: place, EdgeA: edgeA, EdgeB: edgeB, OnEdge: true}
	}

	nodeReach := 18 * scale
	bestID := ""
	bestD := nodeReach
	for id, n := range g.Nodes {
		d := dist(at, pos(n))
		if d <= bestD {
			bestD = d
			bestID = id
		}
	}
	if bestID == "" {
		return mapHit{}
	}
	return mapHit{Place: bestID}
}

// hitPlace keeps the older helper used by tests: place under pointer or nearer road end.
func hitPlace(g *graph.Graph, size fyne.Size, at fyne.Position) string {
	return hitTarget(g, size, at).Place
}

func dist(a, b fyne.Position) float32 {
	return float32(math.Hypot(float64(a.X-b.X), float64(a.Y-b.Y)))
}

func distToSegment(p, a, b fyne.Position) (float32, float32) {
	abx := b.X - a.X
	aby := b.Y - a.Y
	len2 := abx*abx + aby*aby
	if len2 < 1 {
		return dist(p, a), 0
	}
	t := ((p.X-a.X)*abx + (p.Y-a.Y)*aby) / len2
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	proj := fyne.NewPos(a.X+abx*t, a.Y+aby*t)
	return dist(p, proj), t
}
