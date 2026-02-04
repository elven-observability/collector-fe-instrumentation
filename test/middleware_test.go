package test

import (
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "collector-fe-instrumentation/internal/adapter/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := testConfig(t)
	router := gin.New()
	router.Use(httpadapter.JWTAuth(cfg))
	router.GET("/protected/:token", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Success"})
	})

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			token:          generateJWT(jwt.MapClaims{"role": "admin", "iss": "trusted-issuer", "exp": time.Now().Add(1 * time.Hour).Unix()}),
			expectedStatus: 200,
		},
		{
			name:           "Invalid token",
			token:          "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.payload.signature",
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/protected/" + tt.token
			if tt.token == "" {
				path = "/protected/"
			}
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
