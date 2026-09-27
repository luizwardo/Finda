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
	"github.com/wrdo/FInda/internal/protocol"
)

const (
	addrDijkstra = "127.0.0.1:9001"
	addrAStar    = "127.0.0.1:9002"
	animStep     = 55 * time.Millisecond
)

func main() {
	a := app.NewWithID("finda.pathfinder")
	w := a.NewWindow("FInda — Rotas Distribuídas")
	w.Resize(fyne.NewSize(980, 720))

	city := graph.CityMap()
	mapView := NewMapView(city)

	labels, idByLabel := graph.LabelsForSelect(city)
	fromSel := widget.NewSelect(labels, nil)
	toSel := widget.NewSelect(labels, nil)
	if len(labels) >= 2 {
		fromSel.SetSelected(labels[0])
		toSel.SetSelected(labels[len(labels)-1])
		mapView.SetEndpoints(idByLabel[labels[0]], idByLabel[labels[len(labels)-1]])
	}
	fromSel.OnChanged = func(string) {
		mapView.SetEndpoints(idByLabel[fromSel.Selected], idByLabel[toSel.Selected])
	}
	toSel.OnChanged = fromSel.OnChanged

	status := widget.NewLabel("Escolha origem e destino, depois busque a rota nos 2 servidores.")
	status.Wrapping = fyne.TextWrapWord

	legendDijkstra := legendSwatch("Dijkstra (expansão)", colDijkstra)
	legendAStar := legendSwatch("A* (expansão)", colAStar)
	legendBoth := legendSwatch("Ambos exploraram", colBoth)
	legendPath := legendSwatch("Melhor rota", colPath)

	var searching sync.Mutex
	btn := widget.NewButton("Buscar rota (2 servidores)", nil)
	btn.Importance = widget.HighImportance

	btn.OnTapped = func() {
		from := idByLabel[fromSel.Selected]
		to := idByLabel[toSel.Selected]
		if from == "" || to == "" {
			status.SetText("Selecione origem e destino.")
			return
		}
		if from == to {
			status.SetText("Origem e destino devem ser diferentes.")
			return
		}
		if !searching.TryLock() {
			return
		}
		btn.Disable()
		mapView.ResetExploration()
		mapView.SetEndpoints(from, to)
		status.SetText("Consultando servidores Dijkstra (:9001) e A* (:9002)…")

		go func() {
			defer searching.Unlock()
			defer btn.Enable()

			type outcome struct {
				resp protocol.Response
				err  error
				tag  string
			}
			ch := make(chan outcome, 2)
			go func() {
				r, e := queryServer(addrDijkstra, from, to, 3*time.Second)
				ch <- outcome{r, e, "dijkstra"}
			}()
			go func() {
				r, e := queryServer(addrAStar, from, to, 3*time.Second)
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

			if errD != nil && errA != nil {
				fyne.Do(func() {
					status.SetText(fmt.Sprintf("Falha nos dois servidores.\nDijkstra: %v\nA*: %v", errD, errA))
				})
				return
			}

			fyne.Do(func() {
				status.SetText("Animando expansão das buscas…")
			})

			animate(mapView, dijk, astar, errD, errA)

			winner := pickWinner(dijk, astar, errD, errA)
			fyne.Do(func() {
				if winner == nil {
					status.SetText(formatStatus(dijk, astar, errD, errA, nil))
					return
				}
				mapView.SetPath(winner.Path)
				status.SetText(formatStatus(dijk, astar, errD, errA, winner))
			})
		}()
	}

	controls := container.NewVBox(
		widget.NewLabelWithStyle("FInda", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Mapa da cidade · carga distribuída em 2 servidores"),
		container.NewGridWithColumns(2,
			container.NewVBox(widget.NewLabel("Origem"), fromSel),
			container.NewVBox(widget.NewLabel("Destino"), toSel),
		),
		btn,
		container.NewHBox(legendDijkstra, legendAStar, legendBoth, legendPath),
		status,
	)

	w.SetContent(container.NewBorder(
		container.NewPadded(controls),
		nil, nil, nil,
		container.NewPadded(mapView),
	))
	w.ShowAndRun()
}

func legendSwatch(label string, c color.Color) fyne.CanvasObject {
	box := canvas.NewRectangle(c)
	box.SetMinSize(fyne.NewSize(14, 14))
	return container.NewHBox(box, widget.NewLabel(label), layout.NewSpacer())
}

func animate(m *MapView, dijk, astar protocol.Response, errD, errA error) {
	var i, j int
	stepsD := dijk.Steps
	stepsA := astar.Steps
	if errD != nil {
		stepsD = nil
	}
	if errA != nil {
		stepsA = nil
	}

	for i < len(stepsD) || j < len(stepsA) {
		if i < len(stepsD) {
			node := stepsD[i].Node
			fyne.Do(func() { m.MarkDijkstra(node) })
			i++
		}
		if j < len(stepsA) {
			node := stepsA[j].Node
			fyne.Do(func() { m.MarkAStar(node) })
			j++
		}
		time.Sleep(animStep)
	}
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
		return fmt.Sprintf("%s: custo %.1f · %d passos · rota %v", name, r.Cost, len(r.Steps), r.Path)
	}
	msg := line("Dijkstra", dijk, errD) + "\n" + line("A*", astar, errA)
	if winner != nil {
		msg += fmt.Sprintf("\n→ Rota escolhida: %s (custo %.1f)", winner.Algorithm, winner.Cost)
	}
	return msg
}
