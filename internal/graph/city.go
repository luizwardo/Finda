package graph

import (
	"math"
	"sort"
)

// CityMap is the layout drawn by the client. Road lengths here follow the canvas.
func CityMap() *Graph {
	return buildCity(nil, 0)
}

// DijkstraMap is the distance table used by the Dijkstra server.
// The cheap corridor from the airport to the university goes through the port.
func DijkstraMap() *Graph {
	return buildCity(dijkstraWeights, 1)
}

// AStarMap is the distance table used by the A* server.
// The cheap corridor from the airport to the university goes through the arena.
func AStarMap() *Graph {
	return buildCity(astarWeights, 2)
}

func buildCity(weights map[string]float64, salt int) *Graph {
	g := New()

	nodes := []Node{
		{ID: "NP", Name: "Norte Parque", X: 120, Y: 70},
		{ID: "BI", Name: "Biblioteca", X: 320, Y: 60},
		{ID: "HO", Name: "Hospital", X: 540, Y: 80},
		{ID: "TE", Name: "Torre", X: 760, Y: 70},

		{ID: "PC", Name: "Praça Central", X: 160, Y: 200},
		{ID: "SH", Name: "Shopping", X: 360, Y: 210},
		{ID: "ES", Name: "Estação", X: 560, Y: 200},
		{ID: "MU", Name: "Museu", X: 740, Y: 220},

		{ID: "UN", Name: "Universidade", X: 100, Y: 360},
		{ID: "ME", Name: "Mercado", X: 300, Y: 350},
		{ID: "PR", Name: "Porto", X: 520, Y: 370},
		{ID: "BR", Name: "Bridge", X: 700, Y: 360},

		{ID: "ST", Name: "Estádio", X: 140, Y: 520},
		{ID: "AR", Name: "Arena", X: 360, Y: 510},
		{ID: "AE", Name: "Aeroporto", X: 620, Y: 530},
		{ID: "LI", Name: "Litoral", X: 820, Y: 480},

		{ID: "CR", Name: "Correios", X: 40, Y: 140},
		{ID: "FE", Name: "Feira", X: 240, Y: 130},
		{ID: "BC", Name: "Banco", X: 430, Y: 30},
		{ID: "CT", Name: "Centro", X: 450, Y: 140},
		{ID: "JA", Name: "Jardim", X: 660, Y: 140},
		{ID: "ZT", Name: "Zoológico", X: 860, Y: 80},

		{ID: "FA", Name: "Farmácia", X: 40, Y: 250},
		{ID: "IG", Name: "Igreja", X: 230, Y: 275},
		{ID: "TR", Name: "Teatro", X: 450, Y: 275},
		{ID: "CN", Name: "Cinema", X: 660, Y: 275},
		{ID: "DL", Name: "Delegacia", X: 860, Y: 250},

		{ID: "EC", Name: "Escola", X: 40, Y: 430},
		{ID: "RD", Name: "Rodoviária", X: 230, Y: 445},
		{ID: "TM", Name: "Terminal", X: 450, Y: 445},
		{ID: "MR", Name: "Marina", X: 700, Y: 450},
		{ID: "CL", Name: "Clube", X: 860, Y: 400},

		{ID: "VS", Name: "Vista", X: 40, Y: 70},
		{ID: "VT", Name: "VLT", X: 220, Y: 30},
		{ID: "LV", Name: "Livraria", X: 540, Y: 30},
		{ID: "CV", Name: "Conveniência", X: 660, Y: 70},
		{ID: "CA", Name: "Café", X: 120, Y: 140},
		{ID: "PS", Name: "Posto", X: 320, Y: 140},
		{ID: "OF", Name: "Oficina", X: 500, Y: 110},
		{ID: "PT", Name: "Ponto", X: 380, Y: 100},
		{ID: "GA", Name: "Galeria", X: 800, Y: 140},
		{ID: "SJ", Name: "Subestação", X: 860, Y: 180},
		{ID: "HM", Name: "Hotel", X: 800, Y: 180},
		{ID: "PA", Name: "Padaria", X: 240, Y: 200},
		{ID: "PB", Name: "Pub", X: 450, Y: 210},
		{ID: "RF", Name: "Refeitório", X: 660, Y: 210},
		{ID: "UB", Name: "UBS", X: 160, Y: 280},
		{ID: "RS", Name: "Restaurante", X: 360, Y: 280},
		{ID: "LB", Name: "Laboratório", X: 600, Y: 320},
		{ID: "SL", Name: "Salão", X: 760, Y: 300},
		{ID: "BO", Name: "Bosque", X: 40, Y: 330},
		{ID: "BA", Name: "Bairro", X: 200, Y: 380},
		{ID: "AP", Name: "Açougue", X: 400, Y: 360},
		{ID: "MC", Name: "Mecânico", X: 600, Y: 380},
		{ID: "FR", Name: "Floricultura", X: 780, Y: 400},
		{ID: "VL", Name: "Vila", X: 100, Y: 450},
		{ID: "GN", Name: "Ginásio", X: 300, Y: 450},
		{ID: "PI", Name: "Píer", X: 540, Y: 460},
		{ID: "TN", Name: "Tanque", X: 780, Y: 500},
		{ID: "IN", Name: "Indústria", X: 40, Y: 520},
		{ID: "SK", Name: "Skate", X: 200, Y: 510},
		{ID: "SE", Name: "Sede", X: 240, Y: 530},
		{ID: "QK", Name: "Quiosque", X: 500, Y: 540},
		{ID: "TG", Name: "Táxi", X: 860, Y: 520},
	}
	for _, n := range nodes {
		g.AddNode(n)
	}

	roads := [][2]string{
		{"NP", "BI"}, {"BI", "HO"}, {"HO", "TE"},
		{"PC", "SH"}, {"SH", "ES"}, {"ES", "MU"},
		{"UN", "ME"}, {"ME", "PR"}, {"PR", "BR"},
		{"ST", "AR"}, {"AR", "AE"}, {"AE", "LI"},

		{"NP", "PC"}, {"PC", "UN"}, {"UN", "ST"},
		{"BI", "SH"}, {"SH", "ME"}, {"ME", "AR"},
		{"HO", "ES"}, {"ES", "PR"}, {"PR", "AE"},
		{"TE", "MU"}, {"MU", "BR"}, {"BR", "LI"},

		{"PC", "BI"}, {"SH", "HO"}, {"ES", "TE"},
		{"UN", "SH"}, {"ME", "ES"}, {"PR", "MU"},
		{"ST", "ME"}, {"AR", "PR"}, {"AE", "BR"},
		{"PC", "ME"}, {"SH", "PR"}, {"ES", "BR"},

		{"NP", "CR"}, {"CR", "FA"}, {"FA", "EC"}, {"EC", "UN"}, {"EC", "ST"},
		{"NP", "FE"}, {"BI", "FE"}, {"PC", "FE"}, {"SH", "FE"},
		{"BI", "BC"}, {"HO", "BC"}, {"BC", "CT"},
		{"BI", "CT"}, {"HO", "CT"}, {"SH", "CT"}, {"ES", "CT"},
		{"HO", "JA"}, {"TE", "JA"}, {"ES", "JA"}, {"MU", "JA"},
		{"TE", "ZT"}, {"ZT", "DL"}, {"DL", "MU"}, {"DL", "CL"}, {"CL", "BR"}, {"CL", "LI"},
		{"PC", "IG"}, {"SH", "IG"}, {"UN", "IG"}, {"ME", "IG"}, {"FA", "IG"},
		{"SH", "TR"}, {"ES", "TR"}, {"ME", "TR"}, {"PR", "TR"},
		{"ES", "CN"}, {"MU", "CN"}, {"PR", "CN"}, {"BR", "CN"},
		{"UN", "RD"}, {"ME", "RD"}, {"ST", "RD"}, {"AR", "RD"}, {"EC", "RD"},
		{"ME", "TM"}, {"PR", "TM"}, {"AR", "TM"}, {"AE", "TM"},
		{"PR", "MR"}, {"BR", "MR"}, {"AE", "MR"}, {"LI", "MR"},
		{"FE", "CT"}, {"CT", "JA"}, {"JA", "ZT"},
		{"FE", "IG"}, {"IG", "RD"},
		{"CT", "TR"}, {"TR", "TM"},
		{"JA", "CN"}, {"CN", "MR"},
		{"IG", "TR"}, {"TR", "CN"}, {"CN", "DL"},
		{"RD", "TM"}, {"TM", "MR"}, {"MR", "CL"},

		{"VS", "NP"}, {"VS", "CR"},
		{"VT", "NP"}, {"VT", "BI"}, {"VT", "BC"},
		{"LV", "BC"}, {"LV", "HO"},
		{"CV", "HO"}, {"CV", "TE"},
		{"CA", "NP"}, {"CA", "CR"}, {"CA", "PC"},
		{"PS", "BI"}, {"PS", "FE"}, {"PS", "SH"},
		{"OF", "BC"}, {"OF", "CT"}, {"OF", "HO"},
		{"PT", "BI"}, {"PT", "BC"}, {"PT", "CT"},
		{"GA", "TE"}, {"GA", "JA"}, {"GA", "MU"},
		{"SJ", "ZT"}, {"SJ", "DL"}, {"SJ", "TE"},
		{"HM", "MU"}, {"HM", "JA"}, {"HM", "DL"},
		{"PA", "PC"}, {"PA", "FE"}, {"PA", "SH"},
		{"PB", "SH"}, {"PB", "CT"}, {"PB", "ES"},
		{"RF", "ES"}, {"RF", "JA"}, {"RF", "MU"},
		{"UB", "PC"}, {"UB", "IG"}, {"UB", "UN"},
		{"RS", "SH"}, {"RS", "IG"}, {"RS", "TR"}, {"RS", "ME"},
		{"LB", "ES"}, {"LB", "TR"}, {"LB", "PR"}, {"LB", "CN"},
		{"SL", "MU"}, {"SL", "CN"}, {"SL", "BR"},
		{"BO", "FA"}, {"BO", "EC"}, {"BO", "CR"},
		{"BA", "UN"}, {"BA", "ME"}, {"BA", "IG"},
		{"AP", "ME"}, {"AP", "TR"}, {"AP", "PR"},
		{"MC", "PR"}, {"MC", "BR"}, {"MC", "CN"},
		{"FR", "BR"}, {"FR", "CL"}, {"FR", "MU"},
		{"VL", "UN"}, {"VL", "EC"}, {"VL", "ST"}, {"VL", "RD"},
		{"GN", "ME"}, {"GN", "RD"}, {"GN", "AR"}, {"GN", "TM"},
		{"PI", "PR"}, {"PI", "TM"}, {"PI", "AE"}, {"PI", "MR"},
		{"TN", "AE"}, {"TN", "LI"}, {"TN", "MR"},
		{"IN", "ST"}, {"IN", "EC"},
		{"SK", "ST"}, {"SK", "AR"}, {"SK", "SE"},
		{"SE", "ST"}, {"SE", "AR"}, {"SE", "RD"},
		{"QK", "AR"}, {"QK", "AE"}, {"QK", "TM"},
		{"TG", "LI"}, {"TG", "CL"},
	}
	for _, r := range roads {
		w := 0.0
		if weights != nil {
			key := roadKey(r[0], r[1])
			if v, ok := weights[key]; ok {
				w = v
			} else {
				w = fillerWeight(g.Nodes[r[0]], g.Nodes[r[1]], salt)
			}
		}
		g.AddEdge(r[0], r[1], w)
	}
	return g
}

// fillerWeight covers roads that are not in the hand-tuned tables.
// The two salts stay apart so the servers never share a distance.
func fillerWeight(a, b Node, salt int) float64 {
	d := Euclidean(a.X, a.Y, b.X, b.Y)
	n := int(a.X+b.X+a.Y+b.Y) + salt*97
	if n < 0 {
		n = -n
	}
	extra := 28 + (n % 30)
	if salt == 2 {
		extra += 41
	}
	return math.Ceil(d) + float64(extra)
}

func roadKey(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + "-" + b
}

type roadWeight struct {
	a, b string
	w    float64
}

func weightMap(roads []roadWeight) map[string]float64 {
	out := make(map[string]float64, len(roads))
	for _, r := range roads {
		out[roadKey(r.a, r.b)] = r.w
	}
	return out
}

// Distances sit above the straight-line gap on the canvas, so the A* heuristic
// stays admissible. Every road has a different length on the two servers.
var dijkstraWeights = weightMap([]roadWeight{
	{"AE", "BR", 215}, {"AE", "LI", 290}, {"AE", "PR", 192}, {"AR", "AE", 450}, {"AR", "PR", 330},
	{"BI", "HO", 280}, {"BI", "SH", 170}, {"BR", "LI", 205}, {"ES", "BR", 370}, {"ES", "MU", 220},
	{"ES", "PR", 195}, {"ES", "TE", 360}, {"HO", "ES", 140}, {"HO", "TE", 310}, {"ME", "AR", 380},
	{"ME", "ES", 500}, {"ME", "PR", 230}, {"MU", "BR", 165}, {"NP", "BI", 250}, {"NP", "PC", 160},
	{"PC", "BI", 320}, {"PC", "ME", 245}, {"PC", "SH", 240}, {"PC", "UN", 190}, {"PR", "BR", 200},
	{"PR", "MU", 420}, {"SH", "ES", 260}, {"SH", "HO", 340}, {"SH", "ME", 185}, {"SH", "PR", 350},
	{"ST", "AR", 270}, {"ST", "ME", 300}, {"TE", "MU", 175}, {"UN", "ME", 210}, {"UN", "SH", 480},
	{"UN", "ST", 180},
})

var astarWeights = weightMap([]roadWeight{
	{"AE", "BR", 240}, {"AE", "LI", 320}, {"AE", "PR", 510}, {"AR", "AE", 270}, {"AR", "PR", 360},
	{"BI", "HO", 305}, {"BI", "SH", 200}, {"BR", "LI", 250}, {"ES", "BR", 410}, {"ES", "MU", 255},
	{"ES", "PR", 230}, {"ES", "TE", 390}, {"HO", "ES", 155}, {"HO", "TE", 345}, {"ME", "AR", 180},
	{"ME", "ES", 540}, {"ME", "PR", 470}, {"MU", "BR", 190}, {"NP", "BI", 285}, {"NP", "PC", 175},
	{"PC", "BI", 355}, {"PC", "ME", 400}, {"PC", "SH", 275}, {"PC", "UN", 340}, {"PR", "BR", 430},
	{"PR", "MU", 460}, {"SH", "ES", 295}, {"SH", "HO", 375}, {"SH", "ME", 425}, {"SH", "PR", 395},
	{"ST", "AR", 415}, {"ST", "ME", 445}, {"TE", "MU", 210}, {"UN", "ME", 220}, {"UN", "SH", 530},
	{"UN", "ST", 365},
})

// LabelsForSelect returns options for Fyne Select widgets.
func LabelsForSelect(g *Graph) (labels []string, idByLabel map[string]string) {
	ids := g.NodeIDs()
	sort.Strings(ids)
	labels = make([]string, len(ids))
	idByLabel = make(map[string]string, len(ids))
	for i, id := range ids {
		label := g.Nodes[id].Name + "  ·  " + id
		labels[i] = label
		idByLabel[label] = id
	}
	return labels, idByLabel
}
