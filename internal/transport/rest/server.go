package rest

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/robwittman/pillar/internal/service"
)

type ServerConfig struct {
	AgentSvc   service.AgentService
	ConfigSvc  service.ConfigService
	WebhookSvc service.WebhookService
	AttrSvc    service.AttributeService
	LogSvc     *service.LogService
	SourceSvc  service.SourceService
	TriggerSvc service.TriggerService
	TaskSvc    service.TaskService
	AuthSvc    service.AuthService // nil when auth is disabled
	Logger     *slog.Logger
}

func NewServer(cfg ServerConfig) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(slogMiddleware(cfg.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	h := NewHandlers(cfg.AgentSvc, cfg.Logger)
	ch := NewConfigHandlers(cfg.ConfigSvc, cfg.Logger)
	wh := NewWebhookHandlers(cfg.WebhookSvc, cfg.Logger)
	ah := NewAttributeHandlers(cfg.AttrSvc, cfg.Logger)
	lh := NewLogHandlers(cfg.LogSvc, cfg.Logger)
	sh := NewSourceHandlers(cfg.SourceSvc, cfg.Logger)
	trh := NewTriggerHandlers(cfg.TriggerSvc, cfg.Logger)
	tkh := NewTaskHandlers(cfg.TaskSvc, cfg.Logger)

	var authH *AuthHandlers
	authEnabled := cfg.AuthSvc != nil
	if authEnabled {
		authH = NewAuthHandlers(cfg.AuthSvc, cfg.Logger)
	}

	RegisterRoutes(r, h, ch, wh, ah, lh, sh, trh, tkh, authH, cfg.AuthSvc, authEnabled)

	return r
}

func slogMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"duration", time.Since(start),
			)
		})
	}
}
