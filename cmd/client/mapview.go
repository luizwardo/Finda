package main

import (
	"image/color"
	"math"
	"sort"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/pathfind"
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
	dijkTravel  []*travel
	astarTravel []*travel
	pathTravel  *travel
	pathSet     map[string]bool
	pathEdge    map[[2]string]bool
	onHit       func(hit mapHit)
	selEdge     [2]string
	selPulse    float32
	hasSelEdge  bool
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

// SetOnHit registers the callback after a place or road is clicked.
func (m *MapView) SetOnHit(fn func(hit mapHit)) {
	m.onHit = fn
}

// Tapped selects a road (or place) and asks the app which server should leave from there.
func (m *MapView) Tapped(e *fyne.PointEvent) {
	if m.onHit == nil || e == nil {
		return
	}
	hit := m.hitAt(e.Position)
	if hit.Place == "" {
		return
	}
	m.onHit(hit)
}

func (m *MapView) hitAt(at fyne.Position) mapHit {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return hitTarget(m.g, m.Size(), at)
}

// PulseEdge grows a selected road once over ~0.4s.
func (m *MapView) PulseEdge(a, b string) {
	if a == "" || b == "" {
		return
	}
	key := edgeKey(a, b)
	m.mu.Lock()
	m.selEdge = key
	m.hasSelEdge = true
	m.selPulse = 0
	m.mu.Unlock()

	const total = 400 * time.Millisecond
	const frame = 40 * time.Millisecond
	start := time.Now()
	for {
		elapsed := time.Since(start)
		if elapsed >= total {
			break
		}
		t := float32(elapsed) / float32(total)
		m.mu.Lock()
		m.selPulse = easeOutCubic(t)
		m.mu.Unlock()
		fyne.DoAndWait(func() { m.Refresh() })
		time.Sleep(frame)
	}
	m.mu.Lock()
	m.selPulse = 1
	m.mu.Unlock()
	fyne.DoAndWait(func() { m.Refresh() })
}

func easeOutCubic(t float32) float32 {
	u := 1 - clamp01(t)
	return 1 - u*u*u
}

func (m *MapView) ClearEdgePulse() {
	m.mu.Lock()
	m.hasSelEdge = false
	m.selPulse = 0
	m.selEdge = [2]string{}
	m.mu.Unlock()
	m.Refresh()
}

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
	m.hasSelEdge = false
	m.selPulse = 0
	m.selEdge = [2]string{}
	m.mu.Unlock()
	m.Refresh()
}

// RoadLength is the map distance of a road, used to keep the search stroke at a steady speed.
func (m *MapView) RoadLength(from, to string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, okA := m.g.Nodes[from]
	b, okB := m.g.Nodes[to]
	if !okA || !okB {
		return 1
	}
	d := graph.Euclidean(a.X, a.Y, b.X, b.Y)
	if d < 1 {
		return 1
	}
	return d
}

// SetSearchWave commits settled roads and draws every in-flight frontier stroke.
func (m *MapView) SetSearchWave(originD, originA string, doneD, doneA []pathfind.Step, activeD, activeA []activeTravel) {
	m.mu.Lock()
	if originD != "" {
		m.dijk[originD] = true
		m.curDijk = originD
	}
	if originA != "" {
		m.astar[originA] = true
		m.curAStar = originA
	}
	for _, s := range doneD {
		commitStep(m.dijk, m.dijkEdge, &m.curDijk, s)
	}
	for _, s := range doneA {
		commitStep(m.astar, m.astarEdge, &m.curAStar, s)
	}
	m.dijkTravel = travelsFrom(activeD)
	m.astarTravel = travelsFrom(activeA)
	m.mu.Unlock()
	m.Refresh()
}

func travelsFrom(active []activeTravel) []*travel {
	if len(active) == 0 {
		return nil
	}
	out := make([]*travel, 0, len(active))
	for _, a := range active {
		if a.step.From == "" {
			continue
		}
		out = append(out, &travel{from: a.step.From, node: a.step.Node, t: clamp01(a.t)})
	}
	return out
}

func commitStep(seen map[string]bool, edges map[[2]string]bool, current *string, s pathfind.Step) {
	seen[s.Node] = true
	*current = s.Node
	if s.From != "" {
		edges[edgeKey(s.From, s.Node)] = true
	}
}

func clamp01(t float32) float32 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
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

	scale, ox, oy, ok := mapTransform(size)
	if !ok {
		r.objects = objs
		return
	}

	pos := func(n graph.Node) fyne.Position {
		return fyne.NewPos(ox+n.X*scale, oy+n.Y*scale)
	}

	// Stable order: Go maps shuffle each rebuild and made the graph look jumpy.
	type road struct {
		key    [2]string
		a, b   string
		pa, pb fyne.Position
	}
	roads := make([]road, 0, len(r.m.g.Adj)*2)
	seen := make(map[[2]string]bool)
	for from, nbs := range r.m.g.Adj {
		for _, nb := range nbs {
			key := edgeKey(from, nb.ID)
			if seen[key] {
				continue
			}
			seen[key] = true
			// Always draw key[0] → key[1] so blue/orange sides never flip.
			a, b := r.m.g.Nodes[key[0]], r.m.g.Nodes[key[1]]
			roads = append(roads, road{key: key, a: key[0], b: key[1], pa: pos(a), pb: pos(b)})
		}
	}
	sort.Slice(roads, func(i, j int) bool {
		if roads[i].key[0] != roads[j].key[0] {
			return roads[i].key[0] < roads[j].key[0]
		}
		return roads[i].key[1] < roads[j].key[1]
	})

	const side = float32(4.5)
	for _, rd := range roads {
		selected := r.m.hasSelEdge && r.m.selEdge == rd.key
		objs = appendDepthRoad(objs, rd.pa, rd.pb, scale, r.m.selPulse, r.m.pathEdge[rd.key], selected)
		if r.m.dijkEdge[rd.key] {
			objs = append(objs, offsetLine(rd.pa, rd.pb, side*scale, 2*scale, colDijkstra))
		}
		if r.m.astarEdge[rd.key] {
			objs = append(objs, offsetLine(rd.pa, rd.pb, -side*scale, 2*scale, colAStar))
		}
	}

	for _, tr := range r.m.dijkTravel {
		if line := travelLine(r.m.g.Nodes, pos, tr, sideAmount(tr, side)*scale, 2*scale, colDijkstra); line != nil {
			objs = append(objs, line)
		}
	}
	for _, tr := range r.m.astarTravel {
		if line := travelLine(r.m.g.Nodes, pos, tr, sideAmount(tr, -side)*scale, 2*scale, colAStar); line != nil {
			objs = append(objs, line)
		}
	}
	if line := travelLine(r.m.g.Nodes, pos, r.m.pathTravel, 0, 4.5*scale, colPath); line != nil {
		objs = append(objs, line)
	}

	ids := make([]string, 0, len(r.m.g.Nodes))
	for id := range r.m.g.Nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	for _, id := range ids {
		n := r.m.g.Nodes[id]
		p := pos(n)
		radius := float32(6.5) * scale
		fill := colNode
		strong := false
		switch {
		case id == r.m.from:
			fill = colDijkstra
			radius = 9.5 * scale
			strong = true
		case id == r.m.to:
			fill = colAStar
			radius = 9.5 * scale
			strong = true
		case r.m.pathSet[id]:
			fill = colPath
			strong = true
		case r.m.dijk[id] && r.m.astar[id]:
			fill = colBoth
		case r.m.dijk[id]:
			fill = colDijkstra
		case r.m.astar[id]:
			fill = colAStar
		}

		objs = appendDepthNode(objs, n.Name, p, scale, fill, radius, strong)
		if disc := r.growingFill(id, p, radius); disc != nil {
			objs = append(objs, disc)
		}
		if id == r.m.curDijk {
			objs = append(objs, nodeRing(p, radius+5*scale, 2.2*scale, colDijkstra))
		}
		if id == r.m.curAStar {
			objs = append(objs, nodeRing(p, radius+8*scale, 2.2*scale, colAStar))
		}
		for _, tr := range r.m.dijkTravel {
			if dot := travelHead(r.m.g.Nodes, pos, tr, id, sideAmount(tr, side)*scale, 2.5*scale, colDijkstra); dot != nil {
				objs = append(objs, dot)
			}
		}
		for _, tr := range r.m.astarTravel {
			if dot := travelHead(r.m.g.Nodes, pos, tr, id, sideAmount(tr, -side)*scale, 2.5*scale, colAStar); dot != nil {
				objs = append(objs, dot)
			}
		}
		if dot := travelHead(r.m.g.Nodes, pos, r.m.pathTravel, id, 0, 3*scale, colPath); dot != nil {
			objs = append(objs, dot)
		}
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

// sideAmount keeps each algorithm on a fixed geometric side of the road.
// side is relative to the canonical edgeKey(from,to) direction (sorted IDs).
// Traveling against that direction flips the signed offset so the stroke stays put.
func sideAmount(tr *travel, side float32) float32 {
	if tr == nil || tr.from == "" || tr.node == "" {
		return side
	}
	key := edgeKey(tr.from, tr.node)
	if tr.from == key[0] {
		return side
	}
	return -side
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
	if tr == nil || tr.from == "" || tr.t < 0 {
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
	if p := r.m.pathTravel; p != nil && p.node == id {
		return colPath, arrive(p.t), true
	}
	var bestD, bestA float32
	var hitD, hitA bool
	for _, tr := range r.m.dijkTravel {
		if tr != nil && tr.node == id {
			hitD = true
			if tr.t > bestD {
				bestD = tr.t
			}
		}
	}
	for _, tr := range r.m.astarTravel {
		if tr != nil && tr.node == id {
			hitA = true
			if tr.t > bestA {
				bestA = tr.t
			}
		}
	}
	switch {
	case hitD && hitA:
		t := bestD
		if bestA > t {
			t = bestA
		}
		return colBoth, arrive(t), true
	case hitD:
		return colDijkstra, arrive(bestD), true
	case hitA:
		return colAStar, arrive(bestA), true
	default:
		return nil, 0, false
	}
}

// arrive holds the node empty until the stroke is close, then fills it.
// The fill used to occupy the last 45% of the approach; that phase is 25% longer.
func arrive(t float32) float32 {
	const start = float32(0.4375) // 1 - 0.45*1.25
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
