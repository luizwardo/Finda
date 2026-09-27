.PHONY: servers client deps test

servers:
	go run ./cmd/server -algo dijkstra -addr :9001 &
	go run ./cmd/server -algo astar -addr :9002 &
	@echo "Dijkstra :9001  |  A* :9002"

client:
	CGO_ENABLED=1 go run ./cmd/client

test:
	CGO_ENABLED=0 go test ./...

# Fedora/RHEL — run once (needs sudo password in your terminal):
deps:
	sudo dnf install -y gcc libX11-devel libXcursor-devel libXrandr-devel \
		libXinerama-devel libXi-devel libXxf86vm-devel libglvnd-devel mesa-libGL-devel
