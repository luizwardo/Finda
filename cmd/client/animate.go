package main

import (
	"fmt"
	"time"

	"github.com/wrdo/FInda/internal/pathfind"
)

// costPerSec maps path cost to wall-clock so both servers expand as a wavefront.
const costPerSec = 320.0

type activeTravel struct {
	step pathfind.Step
	t    float32
}

// waveAt returns settled nodes and in-flight roads for a search cost radius.
// Several roads can be active at once — the frontier spreads over the map.
func waveAt(m *MapView, steps []pathfind.Step, costNow float64) (origin string, done []pathfind.Step, active []activeTravel) {
	steps = withWaveCosts(m, steps)
	costOf := make(map[string]float64, len(steps))
	for _, s := range steps {
		costOf[s.Node] = s.Cost
	}
	for _, s := range steps {
		if s.From == "" {
			if origin == "" {
				origin = s.Node
			}
			if s.Cost <= costNow {
				done = append(done, s)
			}
			continue
		}
		g0, ok := costOf[s.From]
		if !ok {
			g0 = 0
		}
		g1 := s.Cost
		span := g1 - g0
		if span < 1e-9 {
			span = 1e-9
		}
		switch {
		case costNow >= g1:
			done = append(done, s)
		case costNow > g0:
			t := float32((costNow - g0) / span)
			active = append(active, activeTravel{step: s, t: clamp01(t)})
		}
	}
	return origin, done, active
}

// withWaveCosts fills missing step costs from road lengths so older servers still animate.
func withWaveCosts(m *MapView, steps []pathfind.Step) []pathfind.Step {
	if len(steps) == 0 {
		return steps
	}
	hasCost := false
	for _, s := range steps {
		if s.From != "" && s.Cost > 0 {
			hasCost = true
			break
		}
	}
	if hasCost {
		return steps
	}
	out := make([]pathfind.Step, len(steps))
	copy(out, steps)
	costOf := map[string]float64{}
	for i, s := range out {
		if s.From == "" {
			out[i].Cost = 0
			costOf[s.Node] = 0
			continue
		}
		parent := costOf[s.From]
		L := 120.0
		if m != nil {
			L = m.RoadLength(s.From, s.Node)
		}
		out[i].Cost = parent + L
		costOf[s.Node] = out[i].Cost
	}
	return out
}

func maxStepCost(m *MapView, steps []pathfind.Step) float64 {
	steps = withWaveCosts(m, steps)
	var max float64
	for _, s := range steps {
		if s.Cost > max {
			max = s.Cost
		}
	}
	return max
}

func waveDuration(m *MapView, stepsD, stepsA []pathfind.Step) time.Duration {
	max := maxStepCost(m, stepsD)
	if a := maxStepCost(m, stepsA); a > max {
		max = a
	}
	if max <= 0 {
		return animFrame
	}
	d := time.Duration(max / costPerSec * float64(time.Second))
	if d < 800*time.Millisecond {
		d = 800 * time.Millisecond
	}
	if d > 8*time.Second {
		d = 8 * time.Second
	}
	return d
}

func waveCaption(name string, err error, origin string, steps []pathfind.Step, done []pathfind.Step, active []activeTravel, m *MapView) string {
	if err != nil {
		return fmt.Sprintf("%s: offline (%v)", name, err)
	}
	total := len(steps)
	if total == 0 {
		return name + ": sem passos"
	}
	if len(done) >= total && len(active) == 0 {
		return fmt.Sprintf("%s: concluído · %d nós", name, total)
	}
	here := origin
	if len(active) > 0 {
		here = m.NodeName(active[0].step.Node)
	} else if len(done) > 0 {
		here = m.NodeName(done[len(done)-1].Node)
	}
	return fmt.Sprintf("%s: frente em %d vias · %d/%d nós · perto de %s",
		name, len(active), len(done), total, here)
}
