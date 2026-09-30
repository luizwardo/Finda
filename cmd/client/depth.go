package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	colBG       = color.NRGBA{R: 20, G: 26, B: 38, A: 255}
	colRoad     = color.NRGBA{R: 78, G: 90, B: 112, A: 255}
	colRoadTop  = color.NRGBA{R: 130, G: 144, B: 168, A: 255}
	colRoadSel  = color.NRGBA{R: 96, G: 170, B: 230, A: 255}
	colNode     = color.NRGBA{R: 200, G: 210, B: 226, A: 255}
	colNodeTop  = color.NRGBA{R: 255, G: 255, B: 255, A: 160}
	colShadow   = color.NRGBA{R: 0, G: 0, B: 0, A: 70}
	colDijkstra = color.NRGBA{R: 56, G: 189, B: 248, A: 230}
	colAStar    = color.NRGBA{R: 251, G: 146, B: 60, A: 230}
	colBoth     = color.NRGBA{R: 192, G: 132, B: 252, A: 230}
	colPath     = color.NRGBA{R: 74, G: 222, B: 128, A: 255}
	colLabel    = color.NRGBA{R: 226, G: 232, B: 240, A: 255}
	colLabelBg  = color.NRGBA{R: 14, G: 18, B: 28, A: 190}
)

func appendDepthRoad(objs []fyne.CanvasObject, pa, pb fyne.Position, scale, pulse float32, onPath, selected bool) []fyne.CanvasObject {
	grow := float32(1)
	if selected {
		grow = 1 + 0.55*clamp01(pulse)
	}

	shadow := canvas.NewLine(colShadow)
	shadow.StrokeWidth = 5.5 * scale * grow
	shadow.Position1 = fyne.NewPos(pa.X+1.5*scale, pa.Y+2*scale)
	shadow.Position2 = fyne.NewPos(pb.X+1.5*scale, pb.Y+2*scale)
	objs = append(objs, shadow)

	base := canvas.NewLine(colRoad)
	base.StrokeWidth = 4 * scale * grow
	base.Position1 = pa
	base.Position2 = pb
	switch {
	case onPath:
		base.StrokeColor = colPath
		base.StrokeWidth = 5.5 * scale
	case selected:
		base.StrokeColor = colRoadSel
	}
	objs = append(objs, base)

	hi := canvas.NewLine(colRoadTop)
	hi.StrokeWidth = 1.4 * scale * grow
	hi.Position1 = pa
	hi.Position2 = pb
	if onPath {
		hi.StrokeColor = color.NRGBA{R: 190, G: 255, B: 210, A: 200}
	} else if selected {
		hi.StrokeColor = color.NRGBA{R: 210, G: 235, B: 255, A: 210}
	}
	return append(objs, hi)
}

func appendDepthNode(objs []fyne.CanvasObject, name string, p fyne.Position, scale float32, fill color.Color, radius float32, showLabel bool) []fyne.CanvasObject {
	shadow := canvas.NewCircle(colShadow)
	sr := radius * 1.15
	shadow.Resize(fyne.NewSize(sr*2, sr*1.35))
	shadow.Move(fyne.NewPos(p.X-sr+0.8*scale, p.Y-sr*0.35+1.8*scale))
	objs = append(objs, shadow)

	body := canvas.NewCircle(fill)
	body.Resize(fyne.NewSize(radius*2, radius*2))
	body.Move(fyne.NewPos(p.X-radius, p.Y-radius))
	body.StrokeColor = color.NRGBA{R: 10, G: 14, B: 22, A: 210}
	body.StrokeWidth = 1.4 * scale
	objs = append(objs, body)

	hiR := radius * 0.35
	hi := canvas.NewCircle(colNodeTop)
	hi.Resize(fyne.NewSize(hiR*2, hiR*2))
	hi.Move(fyne.NewPos(p.X-hiR*0.7, p.Y-radius*0.55))
	objs = append(objs, hi)

	if !showLabel {
		return objs
	}
	lw := float32(72) * scale
	lh := float32(12) * scale
	bg := canvas.NewRectangle(colLabelBg)
	bg.CornerRadius = 3 * scale
	bg.Move(fyne.NewPos(p.X-lw/2, p.Y+radius+2*scale))
	bg.Resize(fyne.NewSize(lw, lh))
	objs = append(objs, bg)

	label := canvas.NewText(name, colLabel)
	label.TextSize = 8.5 * scale
	label.TextStyle = fyne.TextStyle{Bold: true}
	label.Alignment = fyne.TextAlignCenter
	label.Move(fyne.NewPos(p.X-lw/2, p.Y+radius+2.2*scale))
	label.Resize(fyne.NewSize(lw, lh))
	return append(objs, label)
}
