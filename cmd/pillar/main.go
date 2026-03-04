package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/robwittman/pillar/internal/config"
	"github.com/robwittman/pillar/internal/runtime"
	"github.com/robwittman/pillar/internal/service"
	pgstore "github.com/robwittman/pillar/internal/storage/postgres"
	redisstore "github.com/robwittman/pillar/internal/storage/redis"
	grpctransport "github.com/robwittman/pillar/internal/transport/grpc"
	"github.com/robwittman/pillar/internal/transport/rest"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logLevel := new(slog.LevelVar)
	switch cfg.LogLevel {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Postgres
	pool, err := pgstore.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Redis
	redisClient, err := redisstore.NewClient(ctx, cfg.RedisAddr)
	if err != nil {
		logger.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	// Repositories
	agentRepo := pgstore.NewAgentRepository(pool)
	statusStore := redisstore.NewAgentStatusStore(redisClient)
	configRepo := pgstore.NewAgentConfigRepository(pool)
	secretStore := pgstore.NewSecretStore(pool)
	integrationRepo := pgstore.NewIntegrationRepository(pool)
	integrationTemplateRepo := pgstore.NewIntegrationTemplateRepository(pool)

	// Stream manager + notifier (shared between service and gRPC layers)
	streamMgr := grpctransport.NewStreamManager()
	notifier := grpctransport.NewStreamNotifier(streamMgr, logger)

	// Services
	svcOpts := []service.AgentServiceOption{service.WithNotifier(notifier)}

	if cfg.KubeEnabled {
		k8sRuntime, err := runtime.NewKubernetesRuntime(runtime.KubernetesConfig{
			Namespace:        cfg.KubeNamespace,
			AgentImage:       cfg.AgentImage,
			GRPCExternalAddr: cfg.GRPCExternalAddr,
		}, logger)
		if err != nil {
			logger.Error("failed to create kubernetes runtime", "error", err)
			os.Exit(1)
		}
		svcOpts = append(svcOpts, service.WithRuntime(k8sRuntime))
		logger.Info("kubernetes runtime enabled", "namespace", cfg.KubeNamespace, "image", cfg.AgentImage)
	}

	integrationTemplateSvc := service.NewIntegrationTemplateService(integrationTemplateRepo, integrationRepo, agentRepo, logger)
	svcOpts = append(svcOpts, service.WithTemplateProvisioner(integrationTemplateSvc))

	agentSvc := service.NewAgentService(agentRepo, statusStore, logger, svcOpts...)
	configSvc := service.NewConfigService(configRepo, agentRepo, secretStore, logger)
	integrationSvc := service.NewIntegrationService(integrationRepo, agentRepo, logger)

	// HTTP server
	httpHandler := rest.NewServer(agentSvc, configSvc, integrationSvc, integrationTemplateSvc, logger)

	mux := http.NewServeMux()
	mux.Handle("/", httpHandler)
	mux.Handle("/metrics", promhttp.Handler())

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	// gRPC server
	grpcServer := grpctransport.NewServer(agentSvc, configSvc, streamMgr, logger)

	// Start servers
	errCh := make(chan error, 2)

	go func() {
		logger.Info("starting HTTP server", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		logger.Info("starting gRPC server", "addr", cfg.GRPCAddr)
		if err := grpctransport.ListenAndServe(grpcServer, cfg.GRPCAddr); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutting down...")
	case err := <-errCh:
		logger.Error("server error", "error", err)
	}

	httpServer.Shutdown(context.Background())
	grpcServer.GracefulStop()
	logger.Info("shutdown complete")
}
