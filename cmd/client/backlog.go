package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"

	"github.com/wrdo/FInda/internal/protocol"
)

const maxBacklog = 4

// searchLog keeps the newest searches at the top of the header backlog.
type searchLog struct {
	mu    sync.Mutex
	items []string
	label *widget.Label
}

func (s *searchLog) prepend(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append([]string{text}, s.items...)
	if len(s.items) > maxBacklog {
		s.items = s.items[:maxBacklog]
	}
	s.label.SetText(strings.Join(s.items, "\n"))
}

func formatBacklog(_ time.Time, _ *MapView, _, _ string, dijk, astar protocol.Response, errD, errA error) string {
	d := parseCost(dijk, errD)
	a := parseCost(astar, errA)
	return fmt.Sprintf("Dijkstra %s · A* %s · Total %s", formatCost(d), formatCost(a), formatTotal(d, a))
}

type maybeCost struct {
	ok    bool
	value float64
}

func parseCost(r protocol.Response, err error) maybeCost {
	if err != nil || !r.OK {
		return maybeCost{}
	}
	return maybeCost{ok: true, value: r.Cost}
}

func formatCost(c maybeCost) string {
	if !c.ok {
		return "—"
	}
	return fmt.Sprintf("%.1f", c.value)
}

func formatTotal(d, a maybeCost) string {
	switch {
	case d.ok && a.ok:
		return fmt.Sprintf("%.1f", d.value+a.value)
	case d.ok:
		return fmt.Sprintf("%.1f", d.value)
	case a.ok:
		return fmt.Sprintf("%.1f", a.value)
	default:
		return "—"
	}
}

func formatElapsed(ns int64) string {
	if ns < 0 {
		ns = 0
	}
	switch {
	case ns < 1_000:
		return fmt.Sprintf("%d ns", ns)
	case ns < 1_000_000:
		return fmt.Sprintf("%.1f µs", float64(ns)/1_000)
	default:
		return fmt.Sprintf("%.2f ms", float64(ns)/1_000_000)
	}
}
