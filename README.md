# Pillar

Pillar is a control plane for headless LLM-powered agents running on remote systems. It manages agent lifecycle, configuration, connectivity, and real-time orchestration over gRPC.

## Architecture

```
                    REST API (:8080)                gRPC Stream (:9090)
                         |                                |
    Operators / CI  ---->|                                |<---- Remote Agents
                         v                                v
                   +-----------+                  +--------------+
                   | HTTP      |                  | gRPC Stream  |
                   | Handlers  |                  | Service      |
                   +-----------+                  +--------------+
                         \                             /
                          \                           /
                           v                         v
                        +---------------------------+
                        |      Agent Service        |
                        |  (CRUD, Start/Stop, HB)   |
                        +---------------------------+
                        /       |         |          \
                       v        v         v           v
                 +---------+ +---------+ +----------+ +----------+
                 | Postgres| |  Redis  | | Notifier | | Runtime  |
                 | (state) | | (online)| | (streams)| | (K8s)    |
                 +---------+ +---------+ +----------+ +----------+
```

- **REST API** -- Operators create agents, attach configuration, and issue start/stop commands.
- **gRPC bidirectional stream** -- Remote agents connect, receive their config and current status, send heartbeats, and receive directives (start/stop) in real-time.
- **Postgres** -- Persistent storage for agents, configs, and secrets.
- **Redis** -- Tracks online/heartbeat status.
- **Runtime (optional)** -- Kubernetes runtime that manages agent Deployments. Opt-in via `PILLAR_KUBE_ENABLED=true`.

## Prerequisites

- Go 1.25+
- Docker & Docker Compose
- `protoc` with `protoc-gen-go` and `protoc-gen-go-grpc` (only if modifying protos)
- [Kind](https://kind.sigs.k8s.io/) (only for Kubernetes runtime)

## Quick Start

### 1. Start Infrastructure

```bash
make docker-up
```

This starts Postgres 16 and Redis 7 via Docker Compose.

### 2. Run Migrations

```bash
make migrate-up
```

### 3. Start the Server

```bash
make run
```

The server starts two listeners:
- **HTTP** on `:8080` (REST API + Prometheus metrics at `/metrics`)
- **gRPC** on `:9090` (agent streams)

### 4. Create and Start an Agent

```bash
# Create an agent
AGENT_ID=$(curl -s localhost:8080/api/v1/agents \
  -H 'Content-Type: application/json' \
  -d '{"name": "my-agent"}' | jq -r .id)

echo "Created agent: $AGENT_ID"

# Attach configuration
curl -s -X POST localhost:8080/api/v1/agents/$AGENT_ID/config \
  -H 'Content-Type: application/json' \
  -d '{
    "model_provider": "claude",
    "model_id": "claude-sonnet-4-20250514",
    "system_prompt": "You are a helpful assistant.",
    "model_params": {"temperature": 0.7, "max_tokens": 4096},
    "max_iterations": 100
  }'

# Start the agent
curl -s -X POST localhost:8080/api/v1/agents/$AGENT_ID/start
```

### 5. Connect an Agent

In a separate terminal, run the example agent:

```bash
make build-example-agent
./bin/example-agent -agent-id $AGENT_ID
```

The example agent will:
- Connect to the gRPC stream and receive its config + current status
- If status is `RUNNING`, log that it would start an LLM loop
- Listen for start/stop directives pushed from the server

### 6. Send Directives

From another terminal, stop and restart the agent:

```bash
# Stop -- the example agent will log the STOP directive
curl -s -X POST localhost:8080/api/v1/agents/$AGENT_ID/stop

# Start again -- the example agent will log the START directive
curl -s -X POST localhost:8080/api/v1/agents/$AGENT_ID/start
```

## Kubernetes Runtime (Optional)

When `PILLAR_KUBE_ENABLED=true`, Start/Stop/Delete commands manage real Kubernetes Deployments in addition to updating the database and pushing gRPC directives.

- **Start** creates a Deployment (replicas=1) or scales an existing one up.
- **Stop** scales the Deployment to 0.
- **Delete** removes the Deployment entirely.

Runtime failures are non-fatal: the DB status update and gRPC directive still succeed even if the K8s API is unreachable.

### Local Setup with Kind

```bash
# Create a Kind cluster (if you don't have one)
kind create cluster

# Build and load the agent image into Kind
make kind-load-agent

# Start infra + server with K8s enabled
make docker-up && make migrate-up
PILLAR_KUBE_ENABLED=true PILLAR_GRPC_EXTERNAL_ADDR=host.docker.internal:9090 make run

# Create, configure, and start an agent
AGENT_ID=$(curl -s localhost:8080/api/v1/agents \
  -H 'Content-Type: application/json' \
  -d '{"name": "k8s-test"}' | jq -r .id)

curl -s -X POST localhost:8080/api/v1/agents/$AGENT_ID/config \
  -H 'Content-Type: application/json' \
  -d '{"model_provider":"claude","model_id":"claude-sonnet-4-20250514","system_prompt":"test"}'

curl -s -X POST localhost:8080/api/v1/agents/$AGENT_ID/start

# Verify the pod is running
kubectl get pods -l pillar.io/agent-id=$AGENT_ID

# Stop → pod terminates
curl -s -X POST localhost:8080/api/v1/agents/$AGENT_ID/stop

# Delete → deployment removed
curl -s -X DELETE localhost:8080/api/v1/agents/$AGENT_ID
```

The agent Pod connects back to Pillar over gRPC using the address from `PILLAR_GRPC_EXTERNAL_ADDR`. With Kind on macOS/Windows, `host.docker.internal:9090` routes to the host machine.

## Configuration

The server is configured via environment variables. Defaults are shown below.

| Variable | Default | Description |
|---|---|---|
| `PILLAR_HTTP_ADDR` | `:8080` | HTTP listen address |
| `PILLAR_GRPC_ADDR` | `:9090` | gRPC listen address |
| `PILLAR_POSTGRES_URL` | `postgres://pillar:pillar@localhost:5432/pillar?sslmode=disable` | Postgres connection string |
| `PILLAR_REDIS_ADDR` | `localhost:6379` | Redis address |
| `PILLAR_LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |
| `PILLAR_KUBE_ENABLED` | `false` | Enable Kubernetes runtime for agent Pods |
| `PILLAR_KUBE_NAMESPACE` | `default` | Namespace for agent Deployments |
| `PILLAR_AGENT_IMAGE` | `pillar-agent:latest` | Container image for agent Pods |
| `PILLAR_GRPC_EXTERNAL_ADDR` | `host.docker.internal:9090` | Address agents use to connect back to Pillar |

A `.env` file in the project root is loaded automatically by most shells. Copy `.env` and adjust as needed.

## REST API

Base path: `/api/v1`

### Health

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Returns `{"status": "ok"}` |

### Agents

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/agents` | Create an agent |
| `GET` | `/api/v1/agents` | List all agents |
| `GET` | `/api/v1/agents/{id}` | Get an agent |
| `PUT` | `/api/v1/agents/{id}` | Update an agent |
| `DELETE` | `/api/v1/agents/{id}` | Delete an agent |
| `POST` | `/api/v1/agents/{id}/start` | Start an agent |
| `POST` | `/api/v1/agents/{id}/stop` | Stop an agent |
| `GET` | `/api/v1/agents/{id}/status` | Get agent status (includes online/offline) |

#### Create Agent

```bash
curl -s localhost:8080/api/v1/agents \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "my-agent",
    "metadata": {"env": "production"},
    "labels": {"team": "infra"}
  }'
```

Response (`201`):

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-agent",
  "status": "pending",
  "metadata": {"env": "production"},
  "labels": {"team": "infra"},
  "created_at": "2026-03-02T10:00:00Z",
  "updated_at": "2026-03-02T10:00:00Z"
}
```

#### Agent Status

```bash
curl -s localhost:8080/api/v1/agents/$AGENT_ID/status
```

Response (`200`):

```json
{
  "agent_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "running",
  "online": true
}
```

`online` is `true` when the agent has an active gRPC connection with recent heartbeats.

#### Start / Stop

Starting an agent transitions it from `pending` or `stopped` to `running`. Stopping transitions from `running` to `stopped`. Invalid transitions return `409 Conflict`.

When an agent is connected over gRPC, start/stop commands push a directive to the agent in real-time.

### Agent Configuration

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/v1/agents/{id}/config` | Create config for an agent |
| `GET` | `/api/v1/agents/{id}/config` | Get config |
| `PUT` | `/api/v1/agents/{id}/config` | Update config |
| `DELETE` | `/api/v1/agents/{id}/config` | Delete config |

#### Create Config

```bash
curl -s -X POST localhost:8080/api/v1/agents/$AGENT_ID/config \
  -H 'Content-Type: application/json' \
  -d '{
    "model_provider": "claude",
    "model_id": "claude-sonnet-4-20250514",
    "system_prompt": "You are a helpful assistant.",
    "api_credential": "sk-ant-...",
    "model_params": {
      "temperature": 0.7,
      "top_p": 0.9,
      "max_tokens": 4096
    },
    "mcp_servers": [
      {
        "name": "filesystem",
        "transport_type": "stdio",
        "command": "mcp-server-filesystem",
        "args": ["/workspace"]
      }
    ],
    "tool_permissions": {
      "allowed_tools": ["read_file", "write_file"],
      "denied_tools": ["execute_command"]
    },
    "max_iterations": 100,
    "token_budget": 100000,
    "task_timeout_seconds": 300,
    "escalation_rules": [
      {
        "name": "error-limit",
        "condition": "error_count > 3",
        "action": "pause",
        "message": "Too many errors, pausing agent"
      }
    ]
  }'
```

The `api_credential` field is stored securely and resolved at connect time. It is never returned in GET responses.

### Error Responses

All errors return JSON:

```json
{"error": "agent not found"}
```

| Status | Meaning |
|---|---|
| `400` | Invalid JSON or missing required fields |
| `404` | Agent or config not found |
| `409` | Invalid state transition or config already exists |
| `500` | Internal server error |

## gRPC Agent Protocol

Agents connect via a bidirectional gRPC stream (`AgentStream` RPC). The protocol flow:

```
Agent                                          Server
  |                                              |
  |--- ConnectRequest {agent_id, capabilities} ->|
  |                                              |
  |<- ConnectAck {accepted, interval,           |
  |               config, status}               |
  |                                              |
  |--- Heartbeat {agent_id, timestamp} -------->|
  |<- HeartbeatAck {server_time} --------------|
  |                                              |
  |    ... (repeats every heartbeat interval) ...|
  |                                              |
  |<- Directive {type: "start"} ----------------|  (when operator calls POST /start)
  |<- Directive {type: "stop"} -----------------|  (when operator calls POST /stop)
  |                                              |
  |--- EventReport {type, payload} ------------>|
  |--- TaskResult {task_id, success, output} --->|
  |                                              |
```

### ConnectAck Fields

| Field | Description |
|---|---|
| `accepted` | Whether the connection was accepted |
| `heartbeat_interval_seconds` | How often the agent should send heartbeats (default: 15s) |
| `config` | Full agent configuration (model, system prompt, MCP servers, etc.) |
| `status` | Current agent status (`PENDING`, `RUNNING`, `STOPPED`, `ERROR`) |

### Directives

Directives are fire-and-forget. The DB status is authoritative -- if an agent reconnects, it gets the correct status in `ConnectAck` regardless of whether it received a directive.

| Type | Meaning |
|---|---|
| `start` | Agent should begin its work loop |
| `stop` | Agent should stop its work loop |

## Client SDK

The Go client SDK (`pkg/client`) provides a high-level interface for agents:

```go
package main

import (
    "context"
    "log/slog"
    "os"

    pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
    "github.com/robwittman/pillar/pkg/client"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

    c, err := client.New("localhost:9090", "my-agent-id", logger)
    if err != nil {
        panic(err)
    }
    defer c.Close()

    // Connect and receive config + status
    if err := c.Connect(context.Background(), nil); err != nil {
        panic(err)
    }

    // Check initial status
    if c.Status() == pillarv1.AgentStatus_AGENT_STATUS_RUNNING {
        // Start work
    }

    // React to directives
    c.OnDirective(func(directiveType, payload string) {
        switch directiveType {
        case "start":
            // Start LLM loop
        case "stop":
            // Stop LLM loop
        }
    })

    // Block and dispatch incoming messages
    c.Listen()
}
```

### Client Methods

| Method | Description |
|---|---|
| `New(addr, agentID, logger)` | Create a new client |
| `Connect(ctx, capabilities)` | Connect to the server, receive config and status, start heartbeat loop |
| `Config()` | Returns the agent config received on connect |
| `Status()` | Returns the current agent status (updated on connect and on directives) |
| `OnDirective(handler)` | Register a callback for incoming directives |
| `Listen()` | Blocking loop that dispatches server messages (heartbeat acks, directives, task assignments) |
| `SendEvent(type, payload)` | Send an event report to the server |
| `SendTaskResult(taskID, success, output, err)` | Send a task result to the server |
| `Close()` | Stop heartbeats and close the connection |

## Makefile Targets

| Target | Description |
|---|---|
| `make build` | Build the server binary to `bin/pillar` |
| `make build-example-agent` | Build the example agent to `bin/example-agent` |
| `make run` | Build and run the server |
| `make test` | Run unit tests with race detector |
| `make test-integration` | Run integration tests (requires Docker) |
| `make docker-up` | Start Postgres and Redis containers |
| `make docker-down` | Stop containers |
| `make migrate-up` | Apply database migrations |
| `make migrate-down` | Roll back database migrations |
| `make proto` | Regenerate protobuf Go code |
| `make docker-build-agent` | Build the agent Docker image (`Dockerfile.agent`) |
| `make kind-load-agent` | Build and load the agent image into a Kind cluster |
| `make lint` | Run golangci-lint |
| `make clean` | Remove build artifacts |

## Project Structure

```
pillar/
  api/proto/pillar/v1/     Protobuf definitions
  cmd/
    pillar/                 Server entrypoint
    example-agent/          Example agent program
  deployments/              Docker Compose files
  gen/proto/pillar/v1/      Generated protobuf Go code
  internal/
    config/                 Server configuration (env vars)
    domain/                 Core types (Agent, AgentConfig, repository interfaces)
    mock/                   Hand-rolled mocks for testing
    runtime/                Agent runtimes (Kubernetes Deployment management)
    service/                Business logic (AgentService, ConfigService)
    storage/
      postgres/             Postgres repositories + migrations
      redis/                Redis status store
    transport/
      grpc/                 gRPC stream service, StreamManager, StreamNotifier
      rest/                 HTTP handlers (Chi router)
  pkg/client/               Go client SDK for agents
  scripts/                  Migration runner
```

```
pillarctl config create 8a972dc2-260a-4817-a9da-43519f7657aa \
    --provider claude \
    --model claude-sonnet-4-6 \
    --api-credential "$ANTHROPIC_API_KEY" \
    --max-iterations 10 \
    --system-prompt "You are a network security engineer. You are tasked with discovering all the resources you can discover from the network you reside in. You can issue read-only type queries, but should abstain from modifying any resources during your test."
```