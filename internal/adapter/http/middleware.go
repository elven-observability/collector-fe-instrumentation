package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"collector-fe-instrumentation/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	paramToken = "token"
)

// JWTAuth returns a Gin middleware that validates JWT from URL param :token.
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	secretKey := []byte(cfg.SecretKey)
	validateExp := cfg.JWTValidateExp
	expectedIssuer := cfg.JWTIssuer
	opts := []jwt.ParserOption{}
	if !validateExp {
		opts = append(opts, jwt.WithoutClaimsValidation())
	}
	parser := jwt.NewParser(opts...)

	return func(c *gin.Context) {
		tokenStr := c.Param(paramToken)
		if tokenStr == "" {
			logAuth(c, "token missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing"})
			c.Abort()
			return
		}

		token, err := parser.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
		if err != nil {
			logAuth(c, fmt.Sprintf("invalid token: %v", err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logAuth(c, "invalid token claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		if !validateRole(c, claims, expectedIssuer) {
			return
		}
		if validateExp && !validateExpClaim(c, claims) {
			return
		}
		c.Next()
	}
}

func validateRole(c *gin.Context, claims jwt.MapClaims, expectedIssuer string) bool {
	role, ok := claims["role"]
	if !ok {
		logAuth(c, "missing role in token")
		c.JSON(http.StatusForbidden, gin.H{"error": "Missing role in token"})
		c.Abort()
		return false
	}
	roleStr := strings.ToLower(fmt.Sprintf("%v", role))
	allowed := []string{"admin", "user"}
	if !contains(allowed, strings.TrimSpace(roleStr)) {
		logAuth(c, fmt.Sprintf("insufficient permissions: role=%s", roleStr))
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
		return false
	}
	// Compare as string (JWT claim can be string or other type)
	if fmt.Sprint(claims["iss"]) != expectedIssuer {
		logAuth(c, fmt.Sprintf("invalid issuer: %v (expected %q)", claims["iss"], expectedIssuer))
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid token issuer", "expected": expectedIssuer})
		c.Abort()
		return false
	}
	return true
}

func validateExpClaim(c *gin.Context, claims jwt.MapClaims) bool {
	exp, ok := claims["exp"].(float64)
	if !ok {
		return true
	}
	if time.Unix(int64(exp), 0).Before(time.Now()) {
		logAuth(c, "token expired")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
		c.Abort()
		return false
	}
	return true
}

func logAuth(c *gin.Context, msg string) {
	slog.Warn("auth",
		"msg", msg,
		"origin", c.Request.Header.Get("Origin"),
		"ip", c.ClientIP(),
		"user_agent", c.Request.UserAgent(),
	)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.TrimSpace(s) == item {
			return true
		}
	}
	return false
}
