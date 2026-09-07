MODULE     := github.com/ibrahimogod/defectdojo-mcp
BINARY     := bin/defectdojo-mcp
IMAGE      := defectdojo-mcp
GHCR_IMAGE := ghcr.io/ibrahimogod/defectdojo-mcp

.PHONY: generate build test lint docker-build refresh-schema

generate:
	go generate ./...

build: generate
	go build -o $(BINARY) ./cmd/server

test: generate
	go test ./...

lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run; else echo "golangci-lint not installed, ran go vet only"; fi

docker-build:
	docker build -t $(IMAGE) .

refresh-schema:
	./scripts/refresh-schema.sh
