.PHONY: build build-ctl build-example-agent run test test-integration lint clean docker-up docker-down migrate-up migrate-down proto docker-build-agent kind-load-agent

BINARY := bin/pillar
AGENT_IMAGE ?= pillar-agent:latest

build:
	go build -o $(BINARY) ./cmd/pillar

build-ctl:
	go build -o bin/pillarctl ./cmd/pillarctl

build-example-agent:
	go build -o bin/example-agent ./cmd/example-agent

run: build
	$(BINARY)

test:
	go test ./... -v -race -count=1

test-integration:
	go test -tags integration ./... -v -race -count=1

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

docker-up:
	docker compose -f deployments/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

migrate-up:
	go run ./scripts/migrate.go -direction up

migrate-down:
	go run ./scripts/migrate.go -direction down

proto:
	protoc --go_out=gen/proto --go_opt=paths=source_relative \
		--go-grpc_out=gen/proto --go-grpc_opt=paths=source_relative \
		-I api/proto api/proto/pillar/v1/agent.proto api/proto/pillar/v1/service.proto

docker-build-agent:
	docker build -f Dockerfile.agent -t $(AGENT_IMAGE) .

kind-load-agent: docker-build-agent
	docker save $(AGENT_IMAGE) | kind load image-archive /dev/stdin
