BIN := "./bin/calendar"
BIN_CS := "./bin/calendar_scheduler"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/calendar_config.toml

build-scheduler:
	go build -v -o $(BIN_CS) -ldflags "$(LDFLAGS)" ./cmd/scheduler

run-scheduler: build-scheduler
	$(BIN_CS) -config ./configs/scheduler_config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

generate:
	rm -rf internal/server/pb/* && \
	protoc \
        --proto_path=api/ \
        --go_out=internal/server/pb \
        --go-grpc_out=internal/server/pb \
        api/*.proto

version: build
	$(BIN) version

test:
	go test -race ./internal/... ./pkg/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build run build-scheduler run-scheduler build-img run-img generate version test lint
