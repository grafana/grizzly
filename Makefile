.PHONY: lint test static install uninstall cross
VERSION := $(shell git describe --tags --dirty --always)
BIN_DIR := $(GOPATH)/bin
GOX := $(BIN_DIR)/gox

lint:
	test -z $$(gofmt -s -l cmd/ pkg/)
	go vet ./...

run-test-image:
	cd pkg/grafana/testdata && docker build . -t grizzly-grafana-test:latest
	docker rm -f grizzly-grafana
	docker run --net $$DRONE_DOCKER_NETWORK_ID --name grizzly-grafana -p 3000:3000 --rm grizzly-grafana-test:latest

run-test-image-locally:
	cd pkg/grafana/testdata && docker build . -t grizzly-grafana-test:latest
	docker rm -f grizzly-grafana
	docker run --name grizzly-grafana -p 3000:3000 --rm grizzly-grafana-test:latest

test:
	go clean -testcache
	make run-test-image-locally &
	go test ./...

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
	go get -u github.com/mitchellh/gox
cross: $(GOX)
	CGO_ENABLED=0 gox -output="dist/{{.Dir}}-{{.OS}}-{{.Arch}}" -ldflags=${LDFLAGS} -arch="amd64 arm64 arm" -os="linux" -osarch="darwin/amd64" ./cmd/grr

# Docker container
container: static
	docker build -t grafana/grizzly .

# CI
drone:
	# Render the YAML from the jsonnet
	drone jsonnet --source .drone/drone.jsonnet --target .drone/drone.yml --stream --format
	# Sign the config
	drone --server https://drone.grafana.net sign --save grafana/grizzly .drone/drone.yml
