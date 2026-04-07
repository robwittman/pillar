package rest

import (
	"github.com/go-chi/chi/v5"
	"github.com/robwittman/pillar/internal/domain"
	"github.com/robwittman/pillar/internal/service"
)

func RegisterRoutes(r chi.Router, h *Handlers, ch *ConfigHandlers, wh *WebhookHandlers, ah *AttributeHandlers, lh *LogHandlers, sh *SourceHandlers, trh *TriggerHandlers, tkh *TaskHandlers, authH *AuthHandlers, orgH *OrgHandlers, authSvc service.AuthService, authEnabled bool, orgRepo domain.OrganizationRepository, membershipRepo domain.MembershipRepository) {
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

		// Auth token routes — no org context needed (tokens are user-scoped).
		if authEnabled && authH != nil {
			r.Route("/auth/tokens", func(r chi.Router) {
				r.Post("/", authH.CreateToken)
				r.Get("/", authH.ListTokens)
				r.Delete("/{tokenID}", authH.RevokeToken)
			})
		}

		// Organization management routes — no org context needed (operate across orgs).
		if orgH != nil {
			r.Route("/organizations", func(r chi.Router) {
				r.Post("/", orgH.CreateOrg)
				r.Get("/", orgH.ListOrgs)
				r.Route("/{orgID}", func(r chi.Router) {
					r.Get("/", orgH.GetOrg)
					r.Put("/", orgH.UpdateOrg)
					r.Delete("/", orgH.DeleteOrg)

					r.Route("/members", func(r chi.Router) {
						r.Get("/", orgH.ListMembers)
						r.Post("/", orgH.AddMember)
						r.Put("/{userID}", orgH.UpdateMemberRole)
						r.Delete("/{userID}", orgH.RemoveMember)
					})

					r.Route("/teams", func(r chi.Router) {
						r.Get("/", orgH.ListTeams)
						r.Post("/", orgH.CreateTeam)
						r.Delete("/{teamID}", orgH.DeleteTeam)
						r.Post("/{teamID}/members", orgH.AddTeamMember)
						r.Delete("/{teamID}/members/{userID}", orgH.RemoveTeamMember)
						r.Get("/{teamID}/members", orgH.ListTeamMembers)
					})
				})
			})
		}

		// Admin routes — no org context needed.
		if authEnabled && authH != nil {
			r.Post("/admin/reconcile-orgs", authH.ReconcilePersonalOrgs)
		}

		// Resource routes — require org context when auth is enabled.
		// Service accounts are org-scoped resources.
		r.Group(func(r chi.Router) {
			if authEnabled && orgRepo != nil && membershipRepo != nil {
				r.Use(OrgResolver(membershipRepo, orgRepo))
			}

			if authEnabled && authH != nil {
				r.Route("/auth/service-accounts", func(r chi.Router) {
					r.Post("/", authH.CreateServiceAccount)
					r.Get("/", authH.ListServiceAccounts)
					r.Delete("/{id}", authH.DeleteServiceAccount)
					r.Post("/{id}/rotate-secret", authH.RotateServiceAccountSecret)
				})
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
		})
	})
}
