
## How to Run

```bash
# dependências Fyne (Fedora) — uma vez
make deps

# terminais separados
go run ./cmd/server -algo dijkstra -addr :9001
go run ./cmd/server -algo astar -addr :9002
CGO_ENABLED=1 go run -tags x11 ./cmd/client
```
