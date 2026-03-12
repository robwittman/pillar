package rest

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handlers, ch *ConfigHandlers, wh *WebhookHandlers, ah *AttributeHandlers, lh *LogHandlers) {
	r.Get("/health", h.Health)

	r.Route("/api/v1", func(r chi.Router) {
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
	})
}
