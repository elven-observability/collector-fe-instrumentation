package domain

import "errors"

var (
	ErrInvalidPayload   = errors.New("invalid payload")
	ErrMissingTenant    = errors.New("tenant or token not provided")
	ErrEmptyPayload     = errors.New("payload has no data")
	ErrLokiSend         = errors.New("failed to send to Loki")
)
