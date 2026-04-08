.PHONY: build build-ctl build-example-agent build-ui dev-ui run test test-integration test-e2e-api test-e2e-ui test-e2e playwright-install lint clean docker-up docker-down migrate-up migrate-down proto docker-build-agent kind-load-agent

BINARY := bin/pillar
AGENT_IMAGE ?= pillar-agent:latest

build-ui:
	cd web && npm install && npm run build

build: build-ui
	go build -o $(BINARY) ./cmd/pillar

build-ctl:
	go build -o bin/pillarctl ./cmd/pillarctl

build-example-agent:
	go build -o bin/example-agent ./cmd/example-agent

build-example-plugin-keycloak:
	go build -o bin/example-plugin-keycloak ./cmd/example-plugin-keycloak

dev-ui:
	cd web && npm run dev

run: build
	$(BINARY)

test:
	go test ./... -v -race -count=1

test-integration:
	go test -tags integration ./... -v -race -count=1

test-e2e-api:
	go test -tags e2e ./tests/e2e/... -v -count=1 -timeout 5m

test-e2e-ui:
	cd tests/playwright && npx playwright test

test-e2e: test-e2e-api test-e2e-ui

playwright-install:
	cd tests/playwright && npm install && npx playwright install chromium

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
	rm -rf web/dist

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
	protoc --go_out=gen/proto --go_opt=paths=source_relative \
		--go-grpc_out=gen/proto --go-grpc_opt=paths=source_relative \
		-I api/proto api/proto/pillar/plugin/v1/plugin.proto

docker-build-agent:
	docker build -f Dockerfile.agent -t $(AGENT_IMAGE) .

kind-load-agent: docker-build-agent
	docker save $(AGENT_IMAGE) | kind load image-archive /dev/stdin
