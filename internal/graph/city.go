package graph

import "sort"

// CityMap returns a fictional downtown map sized for the Fyne canvas (~900x600).
// Layout mimics a small city: park, campus, market, station, port, airport.
func CityMap() *Graph {
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
	}
	for _, n := range nodes {
		g.AddNode(n)
	}

	// Horizontal avenues
	roads := [][2]string{
		{"NP", "BI"}, {"BI", "HO"}, {"HO", "TE"},
		{"PC", "SH"}, {"SH", "ES"}, {"ES", "MU"},
		{"UN", "ME"}, {"ME", "PR"}, {"PR", "BR"},
		{"ST", "AR"}, {"AR", "AE"}, {"AE", "LI"},

		// Verticals
		{"NP", "PC"}, {"PC", "UN"}, {"UN", "ST"},
		{"BI", "SH"}, {"SH", "ME"}, {"ME", "AR"},
		{"HO", "ES"}, {"ES", "PR"}, {"PR", "AE"},
		{"TE", "MU"}, {"MU", "BR"}, {"BR", "LI"},

		// Diagonals / shortcuts (interesting for Dijkstra vs A*)
		{"PC", "BI"}, {"SH", "HO"}, {"ES", "TE"},
		{"UN", "SH"}, {"ME", "ES"}, {"PR", "MU"},
		{"ST", "ME"}, {"AR", "PR"}, {"AE", "BR"},
		{"PC", "ME"}, {"SH", "PR"}, {"ES", "BR"},
	}
	for _, r := range roads {
		g.AddEdge(r[0], r[1], 0)
	}

	return g
}

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
