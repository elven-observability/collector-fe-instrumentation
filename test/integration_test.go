package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpadapter "collector-fe-instrumentation/internal/adapter/http"
	"collector-fe-instrumentation/internal/config"
	"collector-fe-instrumentation/internal/domain"
	"collector-fe-instrumentation/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		t.Fatal("test config invalid:", err)
	}
	return cfg
}

func TestIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testConfig(t)
	router := gin.New()
	router.RedirectTrailingSlash = false
	router.Use(httpadapter.JWTAuth(cfg))
	router.POST("/collect/:tenant/:token", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	tests := []struct {
		name           string
		tenant         string
		token          string
		expectedStatus int
	}{
		{
			name:           "Valid tenant and token",
			tenant:         "elven",
			token:          generateJWT(jwt.MapClaims{"role": "admin", "iss": "trusted-issuer", "exp": time.Now().Add(1 * time.Hour).Unix()}),
			expectedStatus: 200,
		},
		{
			name:           "Invalid token",
			tenant:         "elven",
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.payload.signature",
			expectedStatus: 401,
		},
		{
			name:           "Missing token",
			tenant:         "elven",
			token:          "",
			expectedStatus: 401,
		},
		{
			name:           "Expired token",
			tenant:         "elven",
			token:          generateJWT(jwt.MapClaims{"role": "admin", "iss": "trusted-issuer", "exp": time.Now().Add(-1 * time.Hour).Unix()}),
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/collect/" + tt.tenant + "/" + tt.token
			if tt.token == "" {
				path = "/collect/" + tt.tenant + "/"
			}
			req := httptest.NewRequest(http.MethodPost, path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// noopLoki is a LokiWriter that does nothing (for tests).
type noopLoki struct{}

func (noopLoki) Push(_ context.Context, _ string, _ []domain.LokiStream) error { return nil }

func TestIntegration_CollectRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testConfig(t)
	loki := noopLoki{}
	svc := usecase.NewCollectorService(loki, nil)
	router := httpadapter.Router(cfg, svc)
	token := generateJWT(jwt.MapClaims{"role": "admin", "iss": "trusted-issuer", "exp": time.Now().Add(1 * time.Hour).Unix()})
	path := "/collect/elven/" + token
	body := strings.NewReader(`{"meta":{"app":{"name":"t"},"browser":{},"view":{},"page":{},"session":{},"sdk":{},"user":{}},"logs":[{"message":"hi","level":"info"}]}`)
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
