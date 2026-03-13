package rest

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/robwittman/pillar/internal/service"
)

func NewServer(svc service.AgentService, configSvc service.ConfigService, webhookSvc service.WebhookService, attrSvc service.AttributeService, logSvc *service.LogService, sourceSvc service.SourceService, triggerSvc service.TriggerService, taskSvc service.TaskService, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(slogMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	h := NewHandlers(svc, logger)
	ch := NewConfigHandlers(configSvc, logger)
	wh := NewWebhookHandlers(webhookSvc, logger)
	ah := NewAttributeHandlers(attrSvc, logger)
	lh := NewLogHandlers(logSvc, logger)
	sh := NewSourceHandlers(sourceSvc, logger)
	trh := NewTriggerHandlers(triggerSvc, logger)
	tkh := NewTaskHandlers(taskSvc, logger)
	RegisterRoutes(r, h, ch, wh, ah, lh, sh, trh, tkh)

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
