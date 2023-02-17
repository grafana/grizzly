.PHONY: lint test static install uninstall cross build-test-image run-test-image run-test-image-locally test-clean
VERSION := $(shell git describe --tags --dirty --always)
BIN_DIR := $(GOPATH)/bin
GOX := $(BIN_DIR)/gox

lint:
	test -z $$(gofmt -s -l cmd/ pkg/)
	go vet ./...

build-test-image:
	docker build pkg/grafana/testdata -t grizzly-grafana-test:latest

run-test-image: build-test-image
	docker rm -f grizzly-grafana
	docker run --net $$DRONE_DOCKER_NETWORK_ID --name grizzly-grafana -p 3000:3000 --rm grizzly-grafana-test:latest

run-test-image-locally: build-test-image test-clean
	docker rm -f grizzly-grafana
	docker run -d --name grizzly-grafana -p 3000:3000 --rm grizzly-grafana-test:latest

test-clean:
	go clean -testcache

test: run-test-image-locally
	go test ./... || ( status=$$?; docker logs grizzly-grafana ; exit $$status )

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

# CI
drone:
	# Render the YAML from the jsonnet
	drone jsonnet --source .drone/drone.jsonnet --target .drone/drone.yml --stream --format
	# Sign the config
	drone --server https://drone.grafana.net sign --save grafana/grizzly .drone/drone.yml

