package rest

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handlers, ch *ConfigHandlers) {
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
			})
		})
	})
}
