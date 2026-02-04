package test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpadapter "collector-fe-instrumentation/internal/adapter/http"
	"collector-fe-instrumentation/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestValidateMandatoryTokenClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testConfig(t)
	loki := noopLoki{}
	svc := usecase.NewCollectorService(loki, nil)
	router := httpadapter.Router(cfg, svc)

	tests := []struct {
		name           string
		claims         jwt.MapClaims
		expectedStatus int
	}{
		{
			name: "Valid token claims",
			claims: jwt.MapClaims{
				"role": "admin",
				"iss":  "trusted-issuer",
			},
			expectedStatus: 200,
		},
		{
			name: "Missing role",
			claims: jwt.MapClaims{
				"iss": "trusted-issuer",
			},
			expectedStatus: 403,
		},
		{
			name: "Invalid issuer",
			claims: jwt.MapClaims{
				"role": "admin",
				"iss":  "invalid-issuer",
			},
			expectedStatus: 403,
		},
	}

	minimalPayload := `{"meta":{"app":{"name":"t"},"browser":{},"view":{},"page":{},"session":{},"sdk":{},"user":{}},"logs":[{"message":"ok","level":"info"}]}`
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := generateJWT(tt.claims)
			path := "/collect/tenant/" + token
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(minimalPayload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
