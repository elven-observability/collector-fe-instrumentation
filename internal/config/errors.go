package config

import "errors"

var (
	ErrMissingSecretKey     = errors.New("missing required env: SECRET_KEY")
	ErrSecretKeyTooShort    = errors.New("SECRET_KEY must be at least 64 characters")
	ErrMissingLokiURL       = errors.New("missing required env: LOKI_URL")
	ErrMissingLokiToken     = errors.New("missing required env: LOKI_API_TOKEN")
	ErrMissingAllowOrigins  = errors.New("missing required env: ALLOW_ORIGINS")
)
