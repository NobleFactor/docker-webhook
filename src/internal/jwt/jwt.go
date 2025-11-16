package jwt

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT checks the token using the provided secret
func ValidateJWT(authHeader string, secret []byte) bool {
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return false
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Allow only HMAC methods, with preference for HS512
		switch token.Method {
		case jwt.SigningMethodHS256, jwt.SigningMethodHS384, jwt.SigningMethodHS512:
			return secret, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v (only HMAC methods allowed)", token.Header["alg"])
		}
	})

	if err != nil || !token.Valid {
		return false
	}

	// Optional: check claims (exp, iss)
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().After(time.Unix(int64(exp), 0)) {
				return false
			}
		}
		if iss, ok := claims["iss"].(string); ok {
			if iss != "my-service" {
				return false
			}
		}
	}

	return true
}
