.PHONY: dev lint test integration static install uninstall cross run-test-image-locally stop-test-image-locally test-clean
VERSION := $(shell git describe --tags --dirty --always)
BIN_DIR := $(GOPATH)/bin
GOX := $(BIN_DIR)/gox
DOCKER_COMPOSE := docker compose -f ./test-docker-compose/docker-compose.yml

lint:
	test -z $$(gofmt -s -l cmd/ pkg/)
	go vet ./...

run-test-image-locally: test-clean
	$(DOCKER_COMPOSE) up --force-recreate --detach --remove-orphans --wait

stop-test-image-locally:
	$(DOCKER_COMPOSE) down

test-clean:
	go clean -testcache

test:
	go test -v ./cmd/... ./pkg/...

integration: run-test-image-locally dev
	go test -v ./integration/...
	make stop-test-image-locally

# Compilation
dev:
	go build -ldflags "-X main.Version=dev-${VERSION}" ./cmd/grr

LDFLAGS := '-s -w -extldflags "-static" -X main.Version=${VERSION}'
static:
	CGO_ENABLED=0 GOOS=linux go build -ldflags=${LDFLAGS} ./cmd/grr

install:
	CGO_ENABLED=0 go install -ldflags=${LDFLAGS} ./cmd/grr

uninstall:
	go clean -i ./cmd/grr

$(GOX):
	go install github.com/mitchellh/gox@latest

GOPATH ?= $(HOME)/go
cross: $(GOX)
	CGO_ENABLED=0 $(GOPATH)/bin/gox -output="dist/{{.Dir}}-{{.OS}}-{{.Arch}}" -ldflags=${LDFLAGS} -arch="amd64 arm64 arm" -os="linux" -osarch="darwin/amd64 darwin/arm64" ./cmd/grr

# Docker container
container: static
	docker build -t grafana/grizzly .
