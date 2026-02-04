package config

import (
	"os"
	"strings"
	"time"
)

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

const (
	DefaultHTTPPort    = "3000"
	DefaultLokiTimeout = 15 * time.Second
)

// Config holds application configuration from environment.
type Config struct {
	SecretKey      string
	LokiURL        string
	LokiToken      string
	AllowOrigins   []string
	HTTPPort       string
	LokiTimeout    time.Duration
	JWTValidateExp bool
	JWTIssuer      string
}

// Load reads config from environment.
func Load() *Config {
	origins := getEnv("ALLOW_ORIGINS", "")
	var list []string
	for _, s := range strings.Split(origins, ",") {
		if t := strings.TrimSpace(s); t != "" {
			list = append(list, t)
		}
	}
	validateExp := strings.ToLower(getEnv("JWT_VALIDATE_EXP", "false")) == "true"
	return &Config{
		SecretKey:      getEnv("SECRET_KEY", ""),
		LokiURL:        getEnv("LOKI_URL", ""),
		LokiToken:      getEnv("LOKI_API_TOKEN", ""),
		AllowOrigins:   list,
		HTTPPort:       getEnv("PORT", DefaultHTTPPort),
		LokiTimeout:    DefaultLokiTimeout,
		JWTValidateExp: validateExp,
		JWTIssuer:      getEnv("JWT_ISSUER", "trusted-issuer"),
	}
}

// Validate returns an error if required fields are missing.
func (c *Config) Validate() error {
	if c.SecretKey == "" {
		return ErrMissingSecretKey
	}
	if len(c.SecretKey) < 64 {
		return ErrSecretKeyTooShort
	}
	if c.LokiURL == "" {
		return ErrMissingLokiURL
	}
	if c.LokiToken == "" {
		return ErrMissingLokiToken
	}
	if len(c.AllowOrigins) == 0 {
		return ErrMissingAllowOrigins
	}
	return nil
}
