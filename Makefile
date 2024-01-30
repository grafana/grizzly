.PHONY: dev lint test integration static install uninstall cross build-test-image run-test-image-locally test-clean
VERSION := $(shell git describe --tags --dirty --always)
BIN_DIR := $(GOPATH)/bin
GOX := $(BIN_DIR)/gox

lint:
	test -z $$(gofmt -s -l cmd/ pkg/)
	go vet ./...

build-test-image:
	docker build pkg/grafana/testdata -t grizzly-grafana-test:latest

run-test-image-locally: build-test-image test-clean
	docker rm -f grizzly-grafana
	docker run -d --name grizzly-grafana -p 3001:3001 --rm grizzly-grafana-test:latest

test-clean:
	go clean -testcache

test: run-test-image-locally
	go test -v ./cmd/... ./pkg/... || ( status=$$?; docker logs grizzly-grafana ; exit $$status )

integration: run-test-image-locally dev
	go test -v ./integration/...

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
