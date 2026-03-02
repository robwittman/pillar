package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
	"github.com/robwittman/pillar/pkg/client"
)

func main() {
	addr := flag.String("addr", "localhost:9090", "pillar gRPC address")
	agentID := flag.String("agent-id", "", "agent ID (required)")
	flag.Parse()

	if *agentID == "" {
		fmt.Fprintln(os.Stderr, "error: -agent-id is required")
		os.Exit(1)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	c, err := client.New(*addr, *agentID, logger)
	if err != nil {
		logger.Error("failed to create client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := c.Connect(ctx, nil); err != nil {
		logger.Error("failed to connect", "error", err)
		os.Exit(1)
	}

	if cfg := c.Config(); cfg != nil {
		logger.Info("received config",
			"model_provider", cfg.ModelProvider,
			"model_id", cfg.ModelId,
			"system_prompt_len", len(cfg.SystemPrompt),
		)
	} else {
		logger.Info("no config received")
	}

	status := c.Status()
	logger.Info("initial status", "status", status)
	if status == pillarv1.AgentStatus_AGENT_STATUS_RUNNING {
		logger.Info("status is RUNNING — would start LLM loop here")
	} else {
		logger.Info("status is not RUNNING — idling, waiting for start directive")
	}

	c.OnDirective(func(directiveType, payload string) {
		switch directiveType {
		case "start":
			logger.Info("START directive received — would start LLM loop")
		case "stop":
			logger.Info("STOP directive received — would stop LLM loop")
		default:
			logger.Info("unknown directive", "type", directiveType, "payload", payload)
		}
	})

	listenErr := make(chan error, 1)
	go func() {
		listenErr <- c.Listen()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down...")
	case err := <-listenErr:
		if err != nil {
			logger.Error("listen error", "error", err)
		}
	}
}
