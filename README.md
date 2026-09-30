route-finding app (map-style):** a Fyne GUI client queries **2 servers** (Dijkstra and A*) over sockets/JSON, animates the graph expansion, and highlights the best route
## How to Run

```bash
# dependências Fyne (Fedora) — uma vez
make deps

# terminais separados
go run ./cmd/server -algo dijkstra -addr :9001
go run ./cmd/server -algo astar -addr :9002
CGO_ENABLED=1 go run -tags x11 ./cmd/client
```
