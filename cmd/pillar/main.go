package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/robwittman/pillar/internal/auth"
	"github.com/robwittman/pillar/internal/config"
	"github.com/robwittman/pillar/internal/plugin"
	"github.com/robwittman/pillar/internal/plugin/resolver"
	"github.com/robwittman/pillar/internal/runtime"
	"github.com/robwittman/pillar/internal/service"
	pgstore "github.com/robwittman/pillar/internal/storage/postgres"
	redisstore "github.com/robwittman/pillar/internal/storage/redis"
	grpctransport "github.com/robwittman/pillar/internal/transport/grpc"
	"github.com/robwittman/pillar/internal/transport/rest"
	"github.com/robwittman/pillar/web"
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
	webhookRepo := pgstore.NewWebhookRepository(pool)
	deliveryRepo := pgstore.NewWebhookDeliveryRepository(pool)
	attrRepo := pgstore.NewAgentAttributeRepository(pool)

	// Stream manager + notifier (shared between service and gRPC layers)
	streamMgr := grpctransport.NewStreamManager()
	notifier := grpctransport.NewStreamNotifier(streamMgr, logger)
	attrSvc := service.NewAttributeService(attrRepo, agentRepo, logger)

	// Plugin resolver + manager
	cacheDir := cfg.PluginSettings.CacheDir
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".pillar", "plugins")
	}
	pluginCache := resolver.NewCache(cacheDir, logger)
	pluginResolver := resolver.NewCompositeResolver(pluginCache, logger)

	pluginMgr := plugin.NewManager(logger, plugin.WithResolver(pluginResolver))
	if len(cfg.Plugins) > 0 {
		if err := pluginMgr.StartAll(cfg.Plugins); err != nil {
			logger.Error("failed to start plugins", "error", err)
			os.Exit(1)
		}
		logger.Info("plugins started", "count", len(cfg.Plugins))
	}
	defer pluginMgr.StopAll()

	// Event emitters: plugins (blocking) then webhooks (async)
	pluginEmitter := service.NewPluginEmitter(pluginMgr, attrSvc, logger)
	webhookEmitter := service.NewWebhookEmitter(webhookRepo, deliveryRepo, logger)
	emitter := service.NewCompositeEmitter(pluginEmitter, webhookEmitter)

	svcOpts := []service.AgentServiceOption{
		service.WithNotifier(notifier),
		service.WithEventEmitter(emitter),
	}

	if cfg.KubeEnabled {
		k8sRuntime, err := runtime.NewKubernetesRuntime(runtime.KubernetesConfig{
			Context:          cfg.KubeContext,
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

	agentSvc := service.NewAgentService(agentRepo, statusStore, logger, svcOpts...)
	configSvc := service.NewConfigService(configRepo, agentRepo, secretStore, logger)
	webhookSvc := service.NewWebhookService(webhookRepo, deliveryRepo, logger)

	// Agent log service
	logStore := redisstore.NewAgentLogStore(redisClient)
	logSvc := service.NewLogService(logStore, logger)

	// Task, Source, Trigger services
	sourceRepo := pgstore.NewSourceRepository(pool)
	triggerRepo := pgstore.NewTriggerRepository(pool)
	taskRepo := pgstore.NewTaskRepository(pool)

	taskSvc := service.NewTaskService(taskRepo, notifier, logger)
	triggerSvc := service.NewTriggerService(triggerRepo, logger)
	sourceSvc := service.NewSourceService(sourceRepo, triggerRepo, taskSvc, logger)

	// Webhook worker
	worker := service.NewWebhookWorker(webhookRepo, deliveryRepo, logger)
	worker.Start(ctx)
	defer worker.Stop()

	// Auth (optional)
	var authSvc service.AuthService
	var orgSvc service.OrgService
	var orgRepo *pgstore.OrganizationRepository
	var membershipRepo *pgstore.MembershipRepository
	if cfg.Auth.Enabled {
		userRepo := pgstore.NewUserRepository(pool)
		saRepo := pgstore.NewServiceAccountRepository(pool)
		tokenRepo := pgstore.NewAPITokenRepository(pool)
		sessionStore := redisstore.NewSessionStore(redisClient)
		orgRepo = pgstore.NewOrganizationRepository(pool)
		membershipRepo = pgstore.NewMembershipRepository(pool)

		providerRegistry, err := auth.NewProviderRegistry(ctx, cfg.Auth.Providers, userRepo)
		if err != nil {
			logger.Error("failed to initialize auth providers", "error", err)
			os.Exit(1)
		}

		sessionTTL := 24 * time.Hour
		if parsed, err := time.ParseDuration(cfg.Auth.SessionTTL); err == nil {
			sessionTTL = parsed
		}

		authSvc = service.NewAuthService(
			userRepo, saRepo, tokenRepo, sessionStore,
			orgRepo, membershipRepo,
			providerRegistry, sessionTTL, cfg.Auth.AllowSignup, logger,
		)
		teamRepo := pgstore.NewTeamRepository(pool)
		teamMemberRepo := pgstore.NewTeamMembershipRepository(pool)
		orgSvc = service.NewOrgService(orgRepo, membershipRepo, teamRepo, teamMemberRepo, logger)
		logger.Info("authentication enabled", "providers", len(cfg.Auth.Providers))

		// Bootstrap admin user if local provider is configured and no users exist.
		if providerRegistry.HasLocal() {
			adminEmail := os.Getenv("PILLAR_ADMIN_EMAIL")
			adminPassword := os.Getenv("PILLAR_ADMIN_PASSWORD")
			result, err := auth.Bootstrap(ctx, userRepo, orgRepo, membershipRepo, adminEmail, adminPassword, logger)
			if err != nil {
				logger.Error("failed to bootstrap admin user", "error", err)
				os.Exit(1)
			}
			if result.Created {
				logger.Info("bootstrap: admin user created", "email", result.Email)
				if result.Password != "" {
					fmt.Fprintf(os.Stderr, "\n=== INITIAL ADMIN CREDENTIALS ===\n")
					fmt.Fprintf(os.Stderr, "Email:    %s\n", result.Email)
					fmt.Fprintf(os.Stderr, "Password: %s\n", result.Password)
					fmt.Fprintf(os.Stderr, "=================================\n\n")
				}
			}
		}
	}

	// HTTP server
	httpHandler := rest.NewServer(rest.ServerConfig{
		AgentSvc:       agentSvc,
		ConfigSvc:      configSvc,
		WebhookSvc:     webhookSvc,
		AttrSvc:        attrSvc,
		LogSvc:         logSvc,
		SourceSvc:      sourceSvc,
		TriggerSvc:     triggerSvc,
		TaskSvc:        taskSvc,
		AuthSvc:        authSvc,
		OrgSvc:         orgSvc,
		Logger:         logger,
		OrgRepo:        orgRepo,
		MembershipRepo: membershipRepo,
	})

	// SPA file server from embedded assets
	distFS, _ := fs.Sub(web.Assets, "dist")
	indexHTML, _ := fs.ReadFile(web.Assets, "dist/index.html")
	fileServer := http.FileServer(http.FS(distFS))
	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Serve index.html for root and SPA client-side routes.
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexHTML)
			return
		}
		// Check if static file exists in embedded FS.
		f, err := distFS.Open(strings.TrimPrefix(r.URL.Path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for client-side routes.
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(indexHTML)
	})

	mux := http.NewServeMux()
	mux.Handle("/api/", httpHandler)
	mux.Handle("/auth/", httpHandler)
	mux.Handle("/health", httpHandler)
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", spaHandler)

	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: mux,
	}

	// gRPC server
	grpcServer := grpctransport.NewServer(grpctransport.GRPCServerConfig{
		AgentSvc:       agentSvc,
		ConfigSvc:      configSvc,
		AttrSvc:        attrSvc,
		LogSvc:         logSvc,
		TaskSvc:        taskSvc,
		Streams:        streamMgr,
		Logger:         logger,
		AuthSvc:        authSvc,
		OrgRepo:        orgRepo,
		MembershipRepo: membershipRepo,
	})

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

	_ = httpServer.Shutdown(context.Background())
	grpcServer.GracefulStop()
	logger.Info("shutdown complete")
}
