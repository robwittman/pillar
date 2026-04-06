package rest

import (
	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/service"
)

func RegisterRoutes(r chi.Router, h *Handlers, ch *ConfigHandlers, wh *WebhookHandlers, ah *AttributeHandlers, lh *LogHandlers, sh *SourceHandlers, trh *TriggerHandlers, tkh *TaskHandlers, authH *AuthHandlers, authSvc service.AuthService, authEnabled bool) {
	r.Get("/health", h.Health)

	// Auth endpoints (no auth required).
	if authEnabled && authH != nil {
		r.Route("/auth", func(r chi.Router) {
			r.Get("/providers", authH.ListProviders)
			r.Post("/login", authH.Login)
			r.Post("/register", authH.Register)
			r.Post("/logout", authH.Logout)
			r.Get("/oauth/{provider}", authH.OAuthRedirect)
			r.Get("/oauth/{provider}/callback", authH.OAuthCallback)

			// /auth/me requires auth
			r.Group(func(r chi.Router) {
				r.Use(Authenticator(authSvc))
				r.Get("/me", authH.Me)
			})
		})
	}

	r.Route("/api/v1", func(r chi.Router) {
		// Conditionally apply auth middleware.
		if authEnabled && authSvc != nil {
			r.Use(Authenticator(authSvc))
		}

		r.Route("/agents", func(r chi.Router) {
			r.Post("/", h.CreateAgent)
			r.Get("/", h.ListAgents)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetAgent)
				r.Put("/", h.UpdateAgent)
				r.Delete("/", h.DeleteAgent)
				r.Post("/start", h.StartAgent)
				r.Post("/stop", h.StopAgent)
				r.Get("/status", h.AgentStatus)

				r.Route("/config", func(r chi.Router) {
					r.Post("/", ch.CreateConfig)
					r.Get("/", ch.GetConfig)
					r.Put("/", ch.UpdateConfig)
					r.Delete("/", ch.DeleteConfig)
				})

				r.Route("/attributes", func(r chi.Router) {
					r.Get("/", ah.ListAttributes)
					r.Route("/{namespace}", func(r chi.Router) {
						r.Put("/", ah.SetAttribute)
						r.Get("/", ah.GetAttribute)
						r.Delete("/", ah.DeleteAttribute)
					})
				})

				r.Get("/logs", lh.GetLogs)
				r.Get("/logs/stream", lh.StreamLogs)

				r.Get("/tasks", tkh.ListAgentTasks)
			})
		})

		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/", wh.CreateWebhook)
			r.Get("/", wh.ListWebhooks)
			r.Route("/{webhookID}", func(r chi.Router) {
				r.Get("/", wh.GetWebhook)
				r.Put("/", wh.UpdateWebhook)
				r.Delete("/", wh.DeleteWebhook)
				r.Post("/rotate-secret", wh.RotateSecret)
				r.Get("/deliveries", wh.ListDeliveries)
			})
		})

		r.Route("/sources", func(r chi.Router) {
			r.Post("/", sh.CreateSource)
			r.Get("/", sh.ListSources)
			r.Route("/{sourceID}", func(r chi.Router) {
				r.Get("/", sh.GetSource)
				r.Put("/", sh.UpdateSource)
				r.Delete("/", sh.DeleteSource)
				r.Post("/rotate-secret", sh.RotateSourceSecret)
				r.Post("/events", sh.HandleSourceEvent)
			})
		})

		r.Route("/triggers", func(r chi.Router) {
			r.Post("/", trh.CreateTrigger)
			r.Get("/", trh.ListTriggers)
			r.Route("/{triggerID}", func(r chi.Router) {
				r.Get("/", trh.GetTrigger)
				r.Put("/", trh.UpdateTrigger)
				r.Delete("/", trh.DeleteTrigger)
			})
		})

		r.Route("/tasks", func(r chi.Router) {
			r.Post("/", tkh.CreateTask)
			r.Get("/", tkh.ListTasks)
			r.Route("/{taskID}", func(r chi.Router) {
				r.Get("/", tkh.GetTask)
			})
		})

		// Auth management routes (require auth).
		if authEnabled && authH != nil {
			r.Route("/auth/tokens", func(r chi.Router) {
				r.Post("/", authH.CreateToken)
				r.Get("/", authH.ListTokens)
				r.Delete("/{tokenID}", authH.RevokeToken)
			})
			r.Route("/auth/service-accounts", func(r chi.Router) {
				r.Post("/", authH.CreateServiceAccount)
				r.Get("/", authH.ListServiceAccounts)
				r.Delete("/{id}", authH.DeleteServiceAccount)
				r.Post("/{id}/rotate-secret", authH.RotateServiceAccountSecret)
			})
		}
	})
}
