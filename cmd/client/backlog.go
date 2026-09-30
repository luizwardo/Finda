package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"

	"github.com/wrdo/FInda/internal/protocol"
)

const maxBacklog = 12

// searchLog keeps the newest searches at the top of the side panel.
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
	s.label.SetText(strings.Join(s.items, "\n\n"))
}

func formatBacklog(when time.Time, m *MapView, from, to string, dijk, astar protocol.Response, errD, errA error) string {
	header := fmt.Sprintf("%s   Dijkstra sai de %s · A* sai de %s", when.Format("15:04:05"), m.NodeName(from), m.NodeName(to))
	return header + "\n" +
		backlogServerLine("Dijkstra", m, dijk, errD) + "\n" +
		backlogServerLine("A*", m, astar, errA) + "\n" +
		compareTimes(dijk, astar, errD, errA)
}

func backlogServerLine(name string, m *MapView, r protocol.Response, err error) string {
	if err != nil {
		return fmt.Sprintf("%s: offline (%v)", name, err)
	}
	if !r.OK {
		if r.ElapsedNs > 0 {
			return fmt.Sprintf("%s: %s · %s", name, formatElapsed(r.ElapsedNs), r.Error)
		}
		return fmt.Sprintf("%s: %s", name, r.Error)
	}
	return fmt.Sprintf("%s: %s · custo %.1f · %d nós · %s",
		name, formatElapsed(r.ElapsedNs), r.Cost, len(r.Steps), pathText(m, r.Path))
}

func pathText(m *MapView, ids []string) string {
	if len(ids) == 0 {
		return "—"
	}
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = m.NodeName(id)
	}
	return strings.Join(names, " → ")
}

func compareTimes(dijk, astar protocol.Response, errD, errA error) string {
	okD := errD == nil && dijk.OK && dijk.ElapsedNs > 0
	okA := errA == nil && astar.OK && astar.ElapsedNs > 0
	switch {
	case okD && okA:
		return bothTimes(dijk, astar)
	case okD:
		return "Só o Dijkstra devolveu uma rota (" + formatElapsed(dijk.ElapsedNs) + ")."
	case okA:
		return "Só o A* devolveu uma rota (" + formatElapsed(astar.ElapsedNs) + ")."
	default:
		return "Nenhum servidor devolveu uma rota."
	}
}

func bothTimes(dijk, astar protocol.Response) string {
	cost := "Custo igual."
	if dijk.Cost != astar.Cost {
		cost = fmt.Sprintf("Custos diferentes (Dijkstra %.1f, A* %.1f).", dijk.Cost, astar.Cost)
	}
	if dijk.ElapsedNs == astar.ElapsedNs {
		return fmt.Sprintf("Empate: os dois levaram %s. %s", formatElapsed(dijk.ElapsedNs), cost)
	}
	faster, slower := "A*", "Dijkstra"
	fast, slow := astar.ElapsedNs, dijk.ElapsedNs
	if dijk.ElapsedNs < astar.ElapsedNs {
		faster, slower = "Dijkstra", "A*"
		fast, slow = dijk.ElapsedNs, astar.ElapsedNs
	}
	ratio := float64(slow) / float64(fast)
	return fmt.Sprintf("%s mais rápido: %s contra %s do %s (%.2f×). %s",
		faster, formatElapsed(fast), formatElapsed(slow), slower, ratio, cost)
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
