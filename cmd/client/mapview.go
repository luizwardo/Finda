package main

import (
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"github.com/wrdo/FInda/internal/graph"
)

var (
	colBG       = color.NRGBA{R: 18, G: 24, B: 38, A: 255}
	colRoad     = color.NRGBA{R: 70, G: 85, B: 110, A: 255}
	colNode     = color.NRGBA{R: 200, G: 210, B: 230, A: 255}
	colDijkstra = color.NRGBA{R: 56, G: 189, B: 248, A: 220}  // cyan
	colAStar    = color.NRGBA{R: 251, G: 146, B: 60, A: 220}   // orange
	colBoth     = color.NRGBA{R: 192, G: 132, B: 252, A: 230}  // violet (overlap)
	colPath     = color.NRGBA{R: 74, G: 222, B: 128, A: 255}   // green
	colStart    = color.NRGBA{R: 52, G: 211, B: 153, A: 255}
	colEnd      = color.NRGBA{R: 248, G: 113, B: 113, A: 255}
	colLabel    = color.NRGBA{R: 226, G: 232, B: 240, A: 255}
)

// MapView draws the city graph and animates search expansion.
type MapView struct {
	widget.BaseWidget

	mu       sync.RWMutex
	g        *graph.Graph
	from, to string
	dijk     map[string]bool
	astar    map[string]bool
	pathSet  map[string]bool
	pathEdge map[[2]string]bool
}

func NewMapView(g *graph.Graph) *MapView {
	m := &MapView{
		g:        g,
		dijk:     make(map[string]bool),
		astar:    make(map[string]bool),
		pathSet:  make(map[string]bool),
		pathEdge: make(map[[2]string]bool),
	}
	m.ExtendBaseWidget(m)
	return m
}

func (m *MapView) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(colBG)
	return &mapRenderer{m: m, bg: bg, objects: []fyne.CanvasObject{bg}}
}

func (m *MapView) MinSize() fyne.Size { return fyne.NewSize(900, 580) }

func (m *MapView) SetEndpoints(from, to string) {
	m.mu.Lock()
	m.from, m.to = from, to
	m.mu.Unlock()
	m.Refresh()
}

func (m *MapView) ResetExploration() {
	m.mu.Lock()
	m.dijk = make(map[string]bool)
	m.astar = make(map[string]bool)
	m.pathSet = make(map[string]bool)
	m.pathEdge = make(map[[2]string]bool)
	m.mu.Unlock()
	m.Refresh()
}

func (m *MapView) MarkDijkstra(node string) {
	m.mu.Lock()
	m.dijk[node] = true
	m.mu.Unlock()
	m.Refresh()
}

func (m *MapView) MarkAStar(node string) {
	m.mu.Lock()
	m.astar[node] = true
	m.mu.Unlock()
	m.Refresh()
}

func (m *MapView) SetPath(path []string) {
	m.mu.Lock()
	m.pathSet = make(map[string]bool, len(path))
	m.pathEdge = make(map[[2]string]bool)
	for i, id := range path {
		m.pathSet[id] = true
		if i > 0 {
			m.pathEdge[edgeKey(path[i-1], id)] = true
		}
	}
	m.mu.Unlock()
	m.Refresh()
}

func edgeKey(a, b string) [2]string {
	if a > b {
		a, b = b, a
	}
	return [2]string{a, b}
}

type mapRenderer struct {
	m       *MapView
	bg      *canvas.Rectangle
	objects []fyne.CanvasObject
}

func (r *mapRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.rebuild(size)
}

func (r *mapRenderer) MinSize() fyne.Size { return r.m.MinSize() }

func (r *mapRenderer) Objects() []fyne.CanvasObject { return r.objects }

func (r *mapRenderer) Destroy() {}

func (r *mapRenderer) Refresh() {
	r.bg.FillColor = colBG
	r.bg.Refresh()
	r.rebuild(r.m.Size())
	canvas.Refresh(r.m)
}

func (r *mapRenderer) rebuild(size fyne.Size) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()

	objs := []fyne.CanvasObject{r.bg}
	if r.m.g == nil || size.Width < 1 || size.Height < 1 {
		r.objects = objs
		return
	}

	sx := size.Width / 900
	sy := size.Height / 580
	scale := float32(math.Min(float64(sx), float64(sy)))
	ox := (size.Width - 900*scale) / 2
	oy := (size.Height - 580*scale) / 2

	pos := func(n graph.Node) fyne.Position {
		return fyne.NewPos(ox+n.X*scale, oy+n.Y*scale)
	}

	// Roads
	seen := make(map[[2]string]bool)
	for from, nbs := range r.m.g.Adj {
		for _, nb := range nbs {
			key := edgeKey(from, nb.ID)
			if seen[key] {
				continue
			}
			seen[key] = true
			a, b := r.m.g.Nodes[from], r.m.g.Nodes[nb.ID]
			pa, pb := pos(a), pos(b)
			line := canvas.NewLine(colRoad)
			line.StrokeWidth = 3 * scale
			if r.m.pathEdge[key] {
				line.StrokeColor = colPath
				line.StrokeWidth = 6 * scale
			}
			line.Position1 = pa
			line.Position2 = pb
			objs = append(objs, line)
		}
	}

	// Nodes
	for id, n := range r.m.g.Nodes {
		p := pos(n)
		radius := float32(14) * scale
		fill := colNode
		switch {
		case id == r.m.from:
			fill = colStart
			radius = 18 * scale
		case id == r.m.to:
			fill = colEnd
			radius = 18 * scale
		case r.m.pathSet[id]:
			fill = colPath
		case r.m.dijk[id] && r.m.astar[id]:
			fill = colBoth
		case r.m.dijk[id]:
			fill = colDijkstra
		case r.m.astar[id]:
			fill = colAStar
		}

		circle := canvas.NewCircle(fill)
		circle.Resize(fyne.NewSize(radius*2, radius*2))
		circle.Move(fyne.NewPos(p.X-radius, p.Y-radius))
		circle.StrokeColor = color.NRGBA{R: 15, G: 20, B: 30, A: 255}
		circle.StrokeWidth = 2

		label := canvas.NewText(n.Name, colLabel)
		label.TextSize = 11 * scale
		label.Alignment = fyne.TextAlignCenter
		label.Move(fyne.NewPos(p.X-60*scale, p.Y+radius+2))
		label.Resize(fyne.NewSize(120*scale, 16*scale))

		objs = append(objs, circle, label)
	}

	r.objects = objs
}
