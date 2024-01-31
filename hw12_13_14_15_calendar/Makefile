BIN := "./bin/calendar"
BIN_CS := "./bin/calendar_scheduler"
BIN_SE := "./bin/calendar_sender"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build-calendar:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run-calendar: build
	$(BIN) -config ./configs/calendar_config.toml

build-scheduler:
	go build -v -o $(BIN_CS) -ldflags "$(LDFLAGS)" ./cmd/scheduler

run-scheduler: build-scheduler
	$(BIN_CS) -config ./configs/scheduler_config.toml

build-sender:
	go build -v -o $(BIN_SE) -ldflags "$(LDFLAGS)" ./cmd/sender

run-sender: build-sender
	$(BIN_SE) -config ./configs/sender_config.toml

build: build-calendar build-scheduler build-sender

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

up:
	docker compose -f "./build/docker-compose.yml" up --no-deps --build

down:
	docker compose -f "./build/docker-compose.yml" down

integration-tests:
	docker compose -f "./build/docker-compose.yml" --profile test up -d --no-deps --build
	docker compose -f "./build/docker-compose.yml" --profile test exec -it app-integration-test go test -v ./test/integration_test.go -tags integration
	docker compose -f "./build/docker-compose.yml" --profile test down

generate:
	rm -rf internal/server/pb/* && \
	protoc \
        --proto_path=api/ \
        --go_out=internal/server/pb \
        --go-grpc_out=internal/server/pb \
        api/*.proto

version: build
	$(BIN) version

scheduler-version: build
	$(BIN_CS) version

sender-version: build
	$(BIN) version

test:
	go test -race ./internal/... ./pkg/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2

lint: install-lint-deps
	golangci-lint run ./...

.PHONY: build build-calendar run-calendar build-scheduler run-scheduler build-sender run-sender build-img run-img up down integration-tests generate version scheduler-version sender-version test lint
