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
)
