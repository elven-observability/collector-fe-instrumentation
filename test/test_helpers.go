package test

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func init() {
	_ = os.Setenv("SECRET_KEY", "a-very-secure-key-with-32-characters")
	_ = os.Setenv("JWT_ISSUER", "trusted-issuer")
	_ = os.Setenv("JWT_VALIDATE_EXP", "true")
	_ = os.Setenv("ALLOW_ORIGINS", "http://localhost")
	_ = os.Setenv("LOKI_URL", "http://localhost:3100")
	_ = os.Setenv("LOKI_API_TOKEN", "test-token")
}

func generateJWT(claims jwt.MapClaims) string {
	secretKey := os.Getenv("SECRET_KEY")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		panic("Failed to generate JWT: " + err.Error())
	}
	return tokenString
}
