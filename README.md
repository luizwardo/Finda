# Finda

Aplicação distribuída de busca de rotas (estilo mapa): cliente gráfico em Fyne consulta **2 servidores** (Dijkstra e A*) via sockets/JSON, anima a expansão do grafo e destaca a melhor rota.

## Rodar

```bash
# dependências Fyne (Fedora) — uma vez
make deps

# terminais separados
go run ./cmd/server -algo dijkstra -addr :9001
go run ./cmd/server -algo astar -addr :9002
CGO_ENABLED=1 go run ./cmd/client
```
