package main

import (
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/pathfind"
)

var (
	colBG       = color.NRGBA{R: 18, G: 24, B: 38, A: 255}
	colRoad     = color.NRGBA{R: 70, G: 85, B: 110, A: 255}
	colNode     = color.NRGBA{R: 200, G: 210, B: 230, A: 255}
	colDijkstra = color.NRGBA{R: 56, G: 189, B: 248, A: 220}  // cyan
	colAStar    = color.NRGBA{R: 251, G: 146, B: 60, A: 220}  // orange
	colBoth     = color.NRGBA{R: 192, G: 132, B: 252, A: 230} // violet (overlap)
	colPath     = color.NRGBA{R: 74, G: 222, B: 128, A: 255}  // green
	colStart    = color.NRGBA{R: 52, G: 211, B: 153, A: 255}
	colEnd      = color.NRGBA{R: 248, G: 113, B: 113, A: 255}
	colLabel    = color.NRGBA{R: 226, G: 232, B: 240, A: 255}
)

// MapView draws the city graph and animates search expansion.
type MapView struct {
	widget.BaseWidget

	mu          sync.RWMutex
	g           *graph.Graph
	from, to    string
	dijk        map[string]bool
	astar       map[string]bool
	dijkEdge    map[[2]string]bool
	astarEdge   map[[2]string]bool
	curDijk     string
	curAStar    string
	dijkTravel  *travel
	astarTravel *travel
	pathTravel  *travel
	pathSet     map[string]bool
	pathEdge    map[[2]string]bool
}

// travel is a stroke still moving from one node to the next. t is 0..1.
type travel struct {
	from, node string
	t          float32
}

func NewMapView(g *graph.Graph) *MapView {
	m := &MapView{
		g:         g,
		dijk:      make(map[string]bool),
		astar:     make(map[string]bool),
		dijkEdge:  make(map[[2]string]bool),
		astarEdge: make(map[[2]string]bool),
		pathSet:   make(map[string]bool),
		pathEdge:  make(map[[2]string]bool),
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
	m.dijkEdge = make(map[[2]string]bool)
	m.astarEdge = make(map[[2]string]bool)
	m.curDijk, m.curAStar = "", ""
	m.dijkTravel, m.astarTravel, m.pathTravel = nil, nil, nil
	m.pathSet = make(map[string]bool)
	m.pathEdge = make(map[[2]string]bool)
	m.mu.Unlock()
	m.Refresh()
}

// SetTravels moves each server's stroke along the road it is currently crossing.
// A nil step means that server has already finished. t runs from 0 to 1.
func (m *MapView) SetTravels(dijk, astar *pathfind.Step, t float32) {
	m.mu.Lock()
	m.dijkTravel = travelFrom(dijk, t)
	m.astarTravel = travelFrom(astar, t)
	m.mu.Unlock()
	m.Refresh()
}

// CommitTravels keeps the nodes and roads once the stroke has arrived.
func (m *MapView) CommitTravels() {
	m.mu.Lock()
	commitTravel(m.dijk, m.dijkEdge, &m.curDijk, m.dijkTravel)
	commitTravel(m.astar, m.astarEdge, &m.curAStar, m.astarTravel)
	m.dijkTravel, m.astarTravel = nil, nil
	m.mu.Unlock()
	m.Refresh()
}

func travelFrom(step *pathfind.Step, t float32) *travel {
	if step == nil {
		return nil
	}
	return &travel{from: step.From, node: step.Node, t: smooth(t)}
}

func commitTravel(seen map[string]bool, edges map[[2]string]bool, current *string, tr *travel) {
	if tr == nil {
		return
	}
	seen[tr.node] = true
	*current = tr.node
	if tr.from != "" {
		edges[edgeKey(tr.from, tr.node)] = true
	}
}

func smooth(t float32) float32 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t * t * (3 - 2*t)
}

// BeginRoute clears the previous highlight and marks the start of the chosen route.
func (m *MapView) BeginRoute(start string) {
	m.mu.Lock()
	m.curDijk, m.curAStar = "", ""
	m.pathTravel = nil
	m.pathSet = map[string]bool{start: true}
	m.pathEdge = make(map[[2]string]bool)
	m.mu.Unlock()
	m.Refresh()
}

// SetPathSweep draws the winning route growing from one stop to the next.
func (m *MapView) SetPathSweep(from, to string, t float32) {
	m.mu.Lock()
	m.pathTravel = &travel{from: from, node: to, t: smooth(t)}
	m.mu.Unlock()
	m.Refresh()
}

// CommitPathSweep keeps the road and stop once the green stroke has arrived.
func (m *MapView) CommitPathSweep() {
	m.mu.Lock()
	if m.pathTravel != nil {
		m.pathSet[m.pathTravel.node] = true
		if m.pathTravel.from != "" {
			m.pathEdge[edgeKey(m.pathTravel.from, m.pathTravel.node)] = true
		}
		m.pathTravel = nil
	}
	m.mu.Unlock()
	m.Refresh()
}

func (m *MapView) NodeName(id string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if n, ok := m.g.Nodes[id]; ok && n.Name != "" {
		return n.Name
	}
	return id
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

			// Finished search-tree roads sit beside the street so both algorithms stay visible.
			if r.m.dijkEdge[key] {
				objs = append(objs, offsetLine(pa, pb, 4*scale, 2.5*scale, colDijkstra))
			}
			if r.m.astarEdge[key] {
				objs = append(objs, offsetLine(pa, pb, -4*scale, 2.5*scale, colAStar))
			}
		}
	}

	// Strokes still traveling toward the next node.
	if line := travelLine(r.m.g.Nodes, pos, r.m.dijkTravel, 4*scale, 2.5*scale, colDijkstra); line != nil {
		objs = append(objs, line)
	}
	if line := travelLine(r.m.g.Nodes, pos, r.m.astarTravel, -4*scale, 2.5*scale, colAStar); line != nil {
		objs = append(objs, line)
	}
	if line := travelLine(r.m.g.Nodes, pos, r.m.pathTravel, 0, 6*scale, colPath); line != nil {
		objs = append(objs, line)
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

		objs = append(objs, circle)
		if disc := r.growingFill(id, p, radius); disc != nil {
			objs = append(objs, disc)
		}
		if id == r.m.curDijk {
			objs = append(objs, nodeRing(p, radius+7*scale, 3*scale, colDijkstra))
		}
		if id == r.m.curAStar {
			objs = append(objs, nodeRing(p, radius+12*scale, 3*scale, colAStar))
		}
		if dot := travelHead(r.m.g.Nodes, pos, r.m.dijkTravel, id, 4*scale, 5*scale, colDijkstra); dot != nil {
			objs = append(objs, dot)
		}
		if dot := travelHead(r.m.g.Nodes, pos, r.m.astarTravel, id, -4*scale, 5*scale, colAStar); dot != nil {
			objs = append(objs, dot)
		}
		if dot := travelHead(r.m.g.Nodes, pos, r.m.pathTravel, id, 0, 7*scale, colPath); dot != nil {
			objs = append(objs, dot)
		}
		objs = append(objs, label)
	}

	r.objects = objs
}

func offsetLine(a, b fyne.Position, amount, width float32, c color.Color) *canvas.Line {
	line := canvas.NewLine(c)
	line.StrokeWidth = width
	line.Position1 = offsetPoint(a, b, 0, amount)
	line.Position2 = offsetPoint(a, b, 1, amount)
	return line
}

func offsetPoint(a, b fyne.Position, t, amount float32) fyne.Position {
	end := fyne.NewPos(a.X+(b.X-a.X)*t, a.Y+(b.Y-a.Y)*t)
	dx := float64(b.X - a.X)
	dy := float64(b.Y - a.Y)
	length := math.Hypot(dx, dy)
	if length < 1 {
		return end
	}
	ox := float32(-dy/length) * amount
	oy := float32(dx/length) * amount
	return fyne.NewPos(end.X+ox, end.Y+oy)
}

func travelLine(nodes map[string]graph.Node, pos func(graph.Node) fyne.Position, tr *travel, amount, width float32, c color.Color) *canvas.Line {
	a, b, ok := travelEnds(nodes, pos, tr)
	if !ok {
		return nil
	}
	line := canvas.NewLine(c)
	line.StrokeWidth = width
	line.Position1 = offsetPoint(a, b, 0, amount)
	line.Position2 = offsetPoint(a, b, tr.t, amount)
	return line
}

func travelHead(nodes map[string]graph.Node, pos func(graph.Node) fyne.Position, tr *travel, node string, amount, radius float32, c color.Color) *canvas.Circle {
	if tr == nil || tr.node != node || tr.from == "" {
		return nil
	}
	a, b, ok := travelEnds(nodes, pos, tr)
	if !ok {
		return nil
	}
	at := offsetPoint(a, b, tr.t, amount)
	return disc(at, radius, c)
}

func travelEnds(nodes map[string]graph.Node, pos func(graph.Node) fyne.Position, tr *travel) (fyne.Position, fyne.Position, bool) {
	if tr == nil || tr.from == "" || tr.t <= 0 {
		return fyne.Position{}, fyne.Position{}, false
	}
	a, okA := nodes[tr.from]
	b, okB := nodes[tr.node]
	if !okA || !okB {
		return fyne.Position{}, fyne.Position{}, false
	}
	return pos(a), pos(b), true
}

func (r *mapRenderer) growingFill(id string, center fyne.Position, radius float32) *canvas.Circle {
	if id == r.m.from || id == r.m.to {
		return nil
	}
	c, t, ok := r.fillProgress(id)
	if !ok {
		return nil
	}
	return disc(center, radius*t, c)
}

func (r *mapRenderer) fillProgress(id string) (color.Color, float32, bool) {
	d, a, p := r.m.dijkTravel, r.m.astarTravel, r.m.pathTravel
	dHit := d != nil && d.node == id
	aHit := a != nil && a.node == id
	pHit := p != nil && p.node == id
	switch {
	case pHit:
		return colPath, arrive(p.t), true
	case dHit && aHit:
		t := d.t
		if a.t > t {
			t = a.t
		}
		return colBoth, arrive(t), true
	case dHit:
		return colDijkstra, arrive(d.t), true
	case aHit:
		return colAStar, arrive(a.t), true
	default:
		return nil, 0, false
	}
}

// arrive holds the node empty until the stroke is close, then fills it.
func arrive(t float32) float32 {
	const start = float32(0.55)
	if t <= start {
		return 0
	}
	u := (t - start) / (1 - start)
	return u * u * (3 - 2*u)
}

func disc(center fyne.Position, radius float32, c color.Color) *canvas.Circle {
	if radius < 1 {
		return nil
	}
	circle := canvas.NewCircle(c)
	circle.Resize(fyne.NewSize(radius*2, radius*2))
	circle.Move(fyne.NewPos(center.X-radius, center.Y-radius))
	return circle
}

func nodeRing(center fyne.Position, radius, stroke float32, c color.Color) *canvas.Circle {
	ring := canvas.NewCircle(color.NRGBA{})
	ring.StrokeColor = c
	ring.StrokeWidth = stroke
	ring.Resize(fyne.NewSize(radius*2, radius*2))
	ring.Move(fyne.NewPos(center.X-radius, center.Y-radius))
	return ring
}
