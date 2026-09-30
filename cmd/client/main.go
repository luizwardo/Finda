package main

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/wrdo/FInda/internal/graph"
	"github.com/wrdo/FInda/internal/pathfind"
	"github.com/wrdo/FInda/internal/protocol"
)

const (
	addrDijkstra = "127.0.0.1:9001"
	addrAStar    = "127.0.0.1:9002"
	// Average time to cross one road. Longer roads take longer, so the stroke keeps a steady speed.
	animStep  = 500 * time.Millisecond
	animFrame = 25 * time.Millisecond
)

func main() {
	a := app.NewWithID("finda.pathfinder")
	w := a.NewWindow("FInda — Rotas Distribuídas")
	w.Resize(fyne.NewSize(1180, 760))

	city := graph.CityMap()
	mapView := NewMapView(city)

	ids := city.NodeIDs()
	dijkFrom, astarFrom := ids[0], ids[len(ids)-1]
	mapView.SetEndpoints(dijkFrom, astarFrom)

	dijkName := widget.NewLabel("")
	astarName := widget.NewLabel("")
	status := widget.NewLabel("Clique numa rua (ou num lugar). Depois escolha Dijkstra ou A* no popup.")
	status.Wrapping = fyne.TextWrapWord
	showDepartures := func() {
		dijkName.SetText(departLabel("Dijkstra", mapView, dijkFrom))
		astarName.SetText(departLabel("A*", mapView, astarFrom))
	}
	showDepartures()

	var picking sync.Mutex
	mapView.SetOnHit(func(hit mapHit) {
		if !picking.TryLock() {
			return
		}
		go func() {
			defer picking.Unlock()
			if hit.OnEdge {
				mapView.PulseEdge(hit.EdgeA, hit.EdgeB)
			} else {
				mapView.ClearEdgePulse()
			}

			place := hit.Place
			title := "Saída em " + mapView.NodeName(place)
			detail := "Escolha qual servidor sai daqui."
			if hit.OnEdge {
				detail = "Rua " + mapView.NodeName(hit.EdgeA) + " ↔ " + mapView.NodeName(hit.EdgeB) +
					"\nPonto mais próximo: " + mapView.NodeName(place)
			}

			fyne.DoAndWait(func() {
				body := widget.NewLabel(detail)
				body.Wrapping = fyne.TextWrapWord
				dijkBtn := widget.NewButton("Dijkstra", nil)
				dijkBtn.Importance = widget.HighImportance
				astarBtn := widget.NewButton("A*", nil)
				astarBtn.Importance = widget.WarningImportance
				cancelBtn := widget.NewButton("Cancelar", nil)

				content := container.NewVBox(
					widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
					body,
					container.NewGridWithColumns(2, dijkBtn, astarBtn),
					cancelBtn,
				)
				pop := widget.NewModalPopUp(container.NewPadded(content), w.Canvas())
				pop.Resize(fyne.NewSize(360, 190))

				assign := func(server string) {
					if server == "A*" {
						astarFrom = place
						if dijkFrom == place {
							dijkFrom = ""
						}
					} else {
						dijkFrom = place
						if astarFrom == place {
							astarFrom = ""
						}
					}
					mapView.SetEndpoints(dijkFrom, astarFrom)
					showDepartures()
					status.SetText(departLabel(server, mapView, place) + ". Clique em outra rua para o outro servidor.")
					pop.Hide()
				}
				dijkBtn.OnTapped = func() { assign("Dijkstra") }
				astarBtn.OnTapped = func() { assign("A*") }
				cancelBtn.OnTapped = func() {
					mapView.ClearEdgePulse()
					pop.Hide()
				}
				pop.Show()
			})
		}()
	})

	backlogBody := widget.NewLabel("Nenhuma busca ainda.")
	backlogBody.Wrapping = fyne.TextWrapWord
	backlog := &searchLog{label: backlogBody}

	legendDijkstra := legendSwatch("Dijkstra (expansão)", colDijkstra)
	legendAStar := legendSwatch("A* (expansão)", colAStar)
	legendBoth := legendSwatch("Ambos exploraram", colBoth)
	legendPath := legendSwatch("Melhor rota", colPath)

	var searching sync.Mutex
	btn := widget.NewButton("Encontrar um ao outro", nil)
	btn.Importance = widget.HighImportance

	btn.OnTapped = func() {
		if dijkFrom == "" || astarFrom == "" {
			status.SetText("Clique no mapa para marcar a saída de cada servidor.")
			return
		}
		if dijkFrom == astarFrom {
			status.SetText("Dijkstra e A* precisam sair de lugares diferentes.")
			return
		}
		if !searching.TryLock() {
			return
		}
		btn.Disable()
		mapView.ResetExploration()
		mapView.SetEndpoints(dijkFrom, astarFrom)
		status.SetText("Dijkstra busca o A* e o A* busca o Dijkstra…")

		go func() {
			defer searching.Unlock()
			defer fyne.Do(func() { btn.Enable() })

			type outcome struct {
				resp protocol.Response
				err  error
				tag  string
			}
			ch := make(chan outcome, 2)
			go func() {
				r, e := queryServer(addrDijkstra, dijkFrom, astarFrom, 3*time.Second)
				ch <- outcome{r, e, "dijkstra"}
			}()
			go func() {
				r, e := queryServer(addrAStar, astarFrom, dijkFrom, 3*time.Second)
				ch <- outcome{r, e, "astar"}
			}()

			var dijk, astar protocol.Response
			var errD, errA error
			for i := 0; i < 2; i++ {
				o := <-ch
				if o.tag == "dijkstra" {
					dijk, errD = o.resp, o.err
				} else {
					astar, errA = o.resp, o.err
				}
			}

			entry := formatBacklog(time.Now(), mapView, dijkFrom, astarFrom, dijk, astar, errD, errA)
			if errD != nil && errA != nil {
				fyne.Do(func() {
					status.SetText(fmt.Sprintf("Falha nos dois servidores.\nDijkstra: %v\nA*: %v", errD, errA))
					backlog.prepend(entry)
				})
				return
			}
			fyne.Do(func() { backlog.prepend(entry) })

			fyne.Do(func() {
				status.SetText("Animando expansão das buscas…")
			})

			animate(mapView, status, dijk, astar, errD, errA)

			winner := pickWinner(dijk, astar, errD, errA)
			if winner != nil {
				fyne.Do(func() {
					status.SetText("Desenhando a melhor rota…")
				})
				animateRoute(mapView, winner.Path)
			}
			fyne.Do(func() {
				status.SetText(formatStatus(dijk, astar, errD, errA, winner))
			})
		}()
	}

	controls := container.NewVBox(
		widget.NewLabelWithStyle("FInda", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Clique numa rua; o popup pergunta qual servidor sai dali."),
		container.NewGridWithColumns(2, dijkName, astarName),
		btn,
		container.NewHBox(legendDijkstra, legendAStar, legendBoth, legendPath),
		status,
	)

	backlogScroll := container.NewVScroll(backlogBody)
	backlogPanel := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Backlog", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Tempo de cada um para alcançar o outro."),
		),
		nil, nil, nil,
		backlogScroll,
	)
	split := container.NewHSplit(container.NewPadded(mapView), container.NewPadded(backlogPanel))
	split.SetOffset(0.68)

	w.SetContent(container.NewBorder(
		container.NewPadded(controls),
		nil, nil, nil,
		split,
	))
	w.ShowAndRun()
}

func departLabel(server string, m *MapView, id string) string {
	if id == "" {
		return server + ": clique numa rua"
	}
	return server + " sai de " + m.NodeName(id)
}

func legendSwatch(label string, c color.Color) fyne.CanvasObject {
	box := canvas.NewRectangle(c)
	box.SetMinSize(fyne.NewSize(14, 14))
	return container.NewHBox(box, widget.NewLabel(label), layout.NewSpacer())
}

func animate(m *MapView, status *widget.Label, dijk, astar protocol.Response, errD, errA error) {
	stepsD := dijk.Steps
	stepsA := astar.Steps
	if errD != nil {
		stepsD = nil
	}
	if errA != nil {
		stepsA = nil
	}
	total := waveDuration(m, stepsD, stepsA)
	maxCost := maxStepCost(m, stepsD)
	if a := maxStepCost(m, stepsA); a > maxCost {
		maxCost = a
	}
	rate := costPerSec
	if maxCost > 0 {
		rate = maxCost / total.Seconds()
	}

	var elapsed time.Duration
	for {
		if elapsed > total {
			elapsed = total
		}
		costNow := rate * elapsed.Seconds()
		originD, doneD, activeD := waveAt(m, stepsD, costNow)
		originA, doneA, activeA := waveAt(m, stepsA, costNow)
		text := waveCaption("Dijkstra", errD, originD, stepsD, doneD, activeD, m) + "\n" +
			waveCaption("A*", errA, originA, stepsA, doneA, activeA, m)
		fyne.Do(func() {
			m.SetSearchWave(originD, originA, doneD, doneA, activeD, activeA)
			status.SetText(text)
		})
		if elapsed >= total {
			return
		}
		time.Sleep(animFrame)
		elapsed += animFrame
	}
}

func animFrames() int {
	n := int(animStep / animFrame)
	if n < 1 {
		return 1
	}
	return n
}

func animateRoute(m *MapView, path []string) {
	if len(path) == 0 {
		return
	}
	fyne.Do(func() { m.BeginRoute(path[0]) })
	for i := 1; i < len(path); i++ {
		from, to := path[i-1], path[i]
		frames := animFrames()
		for f := 1; f <= frames; f++ {
			progress := float32(f) / float32(frames)
			fyne.Do(func() { m.SetPathSweep(from, to, progress) })
			time.Sleep(animFrame)
		}
		fyne.Do(func() { m.CommitPathSweep() })
	}
}

func stepCaption(name string, n, total int, step pathfind.Step, m *MapView) string {
	here := m.NodeName(step.Node)
	if step.From == "" {
		return fmt.Sprintf("%s · passo %d/%d · parte de %s", name, n, total, here)
	}
	return fmt.Sprintf("%s · passo %d/%d · %s ← %s", name, n, total, here, m.NodeName(step.From))
}

func pickWinner(dijk, astar protocol.Response, errD, errA error) *protocol.Response {
	okD := errD == nil && dijk.OK && len(dijk.Path) > 0
	okA := errA == nil && astar.OK && len(astar.Path) > 0
	switch {
	case okD && okA:
		if astar.Cost < dijk.Cost {
			return &astar
		}
		return &dijk
	case okD:
		return &dijk
	case okA:
		return &astar
	default:
		return nil
	}
}

func formatStatus(dijk, astar protocol.Response, errD, errA error, winner *protocol.Response) string {
	line := func(name string, r protocol.Response, err error) string {
		if err != nil {
			return fmt.Sprintf("%s: offline (%v)", name, err)
		}
		if !r.OK {
			return fmt.Sprintf("%s: %s", name, r.Error)
		}
		return fmt.Sprintf("%s: %s · custo %.1f · %d passos · rota %v", name, formatElapsed(r.ElapsedNs), r.Cost, len(r.Steps), r.Path)
	}
	msg := line("Dijkstra", dijk, errD) + "\n" + line("A*", astar, errA)
	if winner != nil {
		name := "Dijkstra"
		if winner.Algorithm == protocol.AlgoAStar {
			name = "A*"
		}
		msg += fmt.Sprintf("\n→ Encontro pela rota do %s (custo %.1f)", name, winner.Cost)
	}
	return msg
}
