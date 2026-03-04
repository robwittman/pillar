package rest

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handlers, ch *ConfigHandlers, ih *IntegrationHandlers, ith *IntegrationTemplateHandlers) {
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

				r.Route("/integrations", func(r chi.Router) {
					r.Post("/", ih.CreateIntegration)
					r.Get("/", ih.ListIntegrations)
					r.Route("/{integID}", func(r chi.Router) {
						r.Get("/", ih.GetIntegration)
						r.Put("/", ih.UpdateIntegration)
						r.Delete("/", ih.DeleteIntegration)
					})
				})
			})
		})

		r.Route("/integration-templates", func(r chi.Router) {
			r.Post("/", ith.CreateTemplate)
			r.Get("/", ith.ListTemplates)
			r.Route("/{templateID}", func(r chi.Router) {
				r.Get("/", ith.GetTemplate)
				r.Put("/", ith.UpdateTemplate)
				r.Delete("/", ith.DeleteTemplate)
				r.Get("/preview", ith.PreviewTemplate)
			})
		})
	})
}
