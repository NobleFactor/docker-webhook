// SPDX-FileCopyrightText: 2016-2025 Noble Factor
// SPDX-License-Identifier: MIT
package jwt

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT checks the token using the provided secret and expected location
func ValidateJWT(authHeader string, secretHex string, expectedLocation string) error {
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return fmt.Errorf("missing or empty token")
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

	if err != nil {
		return fmt.Errorf("token parsing failed: %w", err)
	}

	if !token.Valid {
		return fmt.Errorf("token is invalid")
	}

	// Check claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().After(time.Unix(int64(exp), 0)) {
				return fmt.Errorf("token expired at %v", time.Unix(int64(exp), 0))
			}
		}
		if iss, ok := claims["iss"].(string); ok {
			if iss != "webhook-executor" {
				return fmt.Errorf("invalid issuer: expected 'webhook-executor', got '%s'", iss)
			}
		}
		if expectedLocation != "" {
			if sub, ok := claims["sub"].(string); !ok || sub != expectedLocation {
				return fmt.Errorf("invalid subject: expected '%s', got '%s'", expectedLocation, sub)
			}
		}
	} else {
		return fmt.Errorf("missing or invalid claims")
	}

	return nil
}
