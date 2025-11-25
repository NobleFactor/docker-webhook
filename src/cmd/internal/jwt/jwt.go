package jwt

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT checks the token using the provided secret and expected location
func ValidateJWT(authHeader string, secretHex string, expectedLocation string) bool {
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return false
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secretBytes, err := hex.DecodeString(secretHex)
		if err != nil {
			return nil, err
		}
		return secretBytes, nil
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
			if iss != "webhook-executor" {
				return false
			}
		}
		if expectedLocation != "" {
			if sub, ok := claims["sub"].(string); !ok || sub != expectedLocation {
				return false
			}
		}
	}

	return true
}
