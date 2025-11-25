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

// ValidateJWT checks the token using the provided secret and expected location.
//
// Returns: The raw token string and the parsed token on success.
func ValidateJWT(authHeader string, secretHex string, expectedLocation string) (string, *jwt.Token, error) {

	// Parse token

	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenStr == "" {
		return "", nil, fmt.Errorf("invalid authToken: missing or empty value")
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
		return "", nil, fmt.Errorf("invalid authToken: token parsing failed: %w", err)
	}

	if !token.Valid {
		return "", nil, fmt.Errorf("invalid authToken: token is invalid")
	}

	// Check claims

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().After(time.Unix(int64(exp), 0)) {
				return "", nil, fmt.Errorf("invalid authToken: token expired at %v", time.Unix(int64(exp), 0))
			}
		}
		if iss, ok := claims["iss"].(string); ok {
			if iss != "webhook-executor" {
				return "", nil, fmt.Errorf("invalid authToken: invalid issuer: expected 'webhook-executor', got '%s'", iss)
			}
		}
		if expectedLocation != "" {
			if sub, ok := claims["sub"].(string); !ok || sub != expectedLocation {
				return "", nil, fmt.Errorf("invalid authToken: invalid subject: expected '%s', got '%s'", expectedLocation, sub)
			}
		}
	} else {
		return "", nil, fmt.Errorf("invalid authToken: missing or invalid claims")
	}

	return tokenStr, token, nil
}
