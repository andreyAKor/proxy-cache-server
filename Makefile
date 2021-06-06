GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin

build:
	@go build -ldflags="-s -w" -o '$(GOBIN)/ProxyCacheServer' ./cmd/ProxyCacheServer/main.go || exit

run:
	@go build -o '$(GOBIN)/ProxyCacheServer' ./cmd/ProxyCacheServer/main.go
	$(GOBIN)/ProxyCacheServer -config=$(GOBIN)/

test:
	@go test -v -count=1 -race -timeout=60s ./...

install-deps:
	@go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.38.0 && go mod vendor && go mod verify

lint: install-deps
	@golangci-lint run ./...

deps:
	@go mod tidy && go mod vendor && go mod verify

install:
	@go mod download

generate:
	@go generate ./...

.PHONY: build
