package domain

import "errors"

var (
	ErrAgentNotFound      = errors.New("agent not found")
	ErrAgentAlreadyExists = errors.New("agent already exists")
	ErrInvalidTransition  = errors.New("invalid status transition")

	ErrConfigNotFound      = errors.New("agent config not found")
	ErrConfigAlreadyExists = errors.New("agent config already exists")
	ErrInvalidConfig       = errors.New("invalid agent config")
	ErrSecretNotFound      = errors.New("secret not found")

	ErrWebhookNotFound   = errors.New("webhook not found")
	ErrInvalidWebhook    = errors.New("invalid webhook")
	ErrAttributeNotFound = errors.New("attribute not found")

	ErrSourceNotFound  = errors.New("source not found")
	ErrInvalidSource   = errors.New("invalid source")
	ErrTriggerNotFound = errors.New("trigger not found")
	ErrInvalidTrigger  = errors.New("invalid trigger")
	ErrTaskNotFound    = errors.New("task not found")
	ErrInvalidTask     = errors.New("invalid task")

	ErrUserNotFound           = errors.New("user not found")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrServiceAccountNotFound = errors.New("service account not found")
	ErrTokenNotFound          = errors.New("token not found")
	ErrTokenExpired           = errors.New("token expired")
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionExpired         = errors.New("session expired")
	ErrAuthRequired           = errors.New("authentication required")

	ErrOrgNotFound        = errors.New("organization not found")
	ErrOrgAlreadyExists   = errors.New("organization already exists")
	ErrMembershipNotFound = errors.New("membership not found")
	ErrMembershipExists   = errors.New("membership already exists")
	ErrTeamNotFound       = errors.New("team not found")
	ErrNotAuthorized      = errors.New("not authorized")
	ErrOrgContextRequired = errors.New("organization context required")
)
