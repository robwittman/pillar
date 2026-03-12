package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	pillarv1 "github.com/robwittman/pillar/gen/proto/pillar/v1"
	"github.com/robwittman/pillar/pkg/agent"
	"github.com/robwittman/pillar/pkg/client"
)

func main() {
	addr := flag.String("addr", "localhost:9090", "pillar gRPC address")
	agentID := flag.String("agent-id", "", "agent ID (required)")
	task := flag.String("task", "", "one-shot task to run (if empty, waits for directives)")
	streamJSON := flag.Bool("stream-json", false, "emit LLM events as NDJSON to stderr")
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

	cfg := c.Config()
	if cfg != nil {
		logger.Info("received config",
			"model_provider", cfg.ModelProvider,
			"model_id", cfg.ModelId,
			"system_prompt_len", len(cfg.SystemPrompt),
		)
	} else {
		logger.Info("no config received")
	}

	attrs := c.Attributes()
	if len(attrs) > 0 {
		namespaces := make([]string, 0, len(attrs))
		for ns := range attrs {
			namespaces = append(namespaces, ns)
		}
		logger.Info("received attributes", "namespaces", namespaces)
	}

	// Build runner options — always send events to Pillar via gRPC,
	// optionally also write NDJSON to stderr for local debugging.
	grpcEmitter := agent.NewGrpcEmitter(c, *agentID, logger)
	var emitter agent.Emitter = grpcEmitter
	if *streamJSON {
		emitter = agent.NewMultiEmitter(grpcEmitter, agent.NewEventWriter(os.Stderr, *agentID))
	}
	runnerOpts := []agent.RunnerOption{agent.WithEvents(emitter)}

	// One-shot mode: run a task immediately and exit
	if *task != "" {
		if cfg == nil {
			logger.Error("no config received, cannot run task")
			os.Exit(1)
		}
		runTask(ctx, cfg, attrs, *task, logger, c, runnerOpts)
		return
	}

	// Directive mode: wait for start/stop
	var (
		runnerCancel context.CancelFunc
		runnerMu     sync.Mutex
	)

	c.OnDirective(func(directiveType, payload string) {
		runnerMu.Lock()
		defer runnerMu.Unlock()

		switch directiveType {
		case "start":
			logger.Info("START directive received")
			if runnerCancel != nil {
				return // already running
			}
			if cfg == nil {
				logger.Error("no config, cannot start")
				return
			}
			runCtx, runCancel := context.WithCancel(ctx)
			runnerCancel = runCancel

			taskPrompt := payload
			if taskPrompt == "" {
				taskPrompt = "You are now active. Follow your system prompt instructions."
			}

			go func() {
				runTask(runCtx, cfg, attrs, taskPrompt, logger, c, runnerOpts)
				runnerMu.Lock()
				runnerCancel = nil
				runnerMu.Unlock()
			}()

		case "stop":
			logger.Info("STOP directive received")
			if runnerCancel != nil {
				runnerCancel()
				runnerCancel = nil
			}
		}
	})

	status := c.Status()
	if status == pillarv1.AgentStatus_AGENT_STATUS_RUNNING && cfg != nil {
		logger.Info("status is RUNNING — starting agent loop")
		runCtx, runCancel := context.WithCancel(ctx)
		runnerCancel = runCancel
		go func() {
			runTask(runCtx, cfg, attrs, "You are now active. Follow your system prompt instructions.", logger, c, runnerOpts)
			runnerMu.Lock()
			runnerCancel = nil
			runnerMu.Unlock()
		}()
	} else {
		logger.Info("waiting for start directive", "status", status)
	}

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

func runTask(ctx context.Context, cfg *pillarv1.AgentConfig, attrs map[string][]byte, task string, logger *slog.Logger, c *client.Client, opts []agent.RunnerOption) {
	runner, err := agent.NewRunner(cfg, attrs, logger, opts...)
	if err != nil {
		logger.Error("failed to create runner", "error", err)
		c.SendEvent("error", fmt.Sprintf("failed to create runner: %s", err))
		return
	}

	logger.Info("starting agent task", "task_len", len(task))
	c.SendEvent("task.started", task)

	result, err := runner.Run(ctx, task)
	if err != nil {
		logger.Error("agent task failed", "error", err)
		c.SendEvent("task.failed", err.Error())
		return
	}

	logger.Info("agent task completed", "result_len", len(result))
	fmt.Println("\n--- Agent Result ---")
	fmt.Println(result)
	fmt.Println("--- End Result ---")
	c.SendEvent("task.completed", result)
}
