// SPDX-FileCopyrightText: 2016-2025 Noble Factor
// SPDX-License-Identifier: MIT
package jwt

import (
    "encoding/hex"
    "fmt"
    "strings"
    "time"
    "strconv"

    "github.com/golang-jwt/jwt/v5"
)

// RefreshJWT refreshes the incoming Authorization header token using the provided secret (hex-encoded).
// If the token is within tokenRefreshWindow of expiry (or already expired) it returns a newly signed token with
// tokenTtl. The function mirrors the original token's "alg" header for signing when possible.
//
// Returns: (newToken, refreshed, error)
// RefreshJWT refreshes the provided parsed token using the provided secret (hex-encoded).
// The parsed token must already have been validated (signature + claims) by `ValidateJWT`.
// If the token is within tokenRefreshWindow of expiry (or already expired) it returns a newly signed token with
// tokenTtl. The function mirrors the original token's "alg" header for signing when possible.

// Returns: (newToken, refreshed, error)
func RefreshJWT(parsed *jwt.Token, tokenStr string, secretHex string, expectedLocation string, tokenRefreshWindow, tokenTtl time.Duration) (string, bool, error) {

    if parsed == nil {
        return "", false, fmt.Errorf("parsed token is nil")
    }

    // Decode secret
    secretBytes, err := hex.DecodeString(secretHex)
    if err != nil {
        return "", false, fmt.Errorf("invalid secret hex: %w", err)
    }

    // Extract claims
    claims, ok := parsed.Claims.(jwt.MapClaims)
    if !ok {
        return "", false, fmt.Errorf("token has invalid claims type")
    }

    // verify subject if required
    if expectedLocation != "" {
        if sub, ok := claims["sub"].(string); !ok || sub != expectedLocation {
            return "", false, fmt.Errorf("invalid subject: expected %q", expectedLocation)
        }
    }

    // determine algorithm from token header (fallback to HS512)
    alg := "HS512"
    if parsed.Method != nil {
        alg = parsed.Method.Alg()
    } else if v, ok := parsed.Header["alg"].(string); ok && v != "" {
        alg = v
    }
    alg = strings.ToUpper(alg)

    // compute remaining TTL
    now := time.Now()
    var expTime time.Time
    if expRaw, ok := claims["exp"]; ok {
        switch v := expRaw.(type) {
        case float64:
            expTime = time.Unix(int64(v), 0)
        case int64:
            expTime = time.Unix(v, 0)
        case string:
            if i, err := strconv.ParseInt(v, 10, 64); err == nil {
                expTime = time.Unix(i, 0)
            }
        default:
            // leave zero time
        }
    }

    ttlRemaining := time.Duration(0)
    if !expTime.IsZero() {
        ttlRemaining = expTime.Sub(now)
    } else {
        // no exp claim -> treat as needing refresh
        ttlRemaining = -1
    }

    // decide whether to refresh: if no exp or ttlRemaining <= tokenRefreshWindow
    if !expTime.IsZero() && ttlRemaining > tokenRefreshWindow {
        // no refresh needed; return the original token so callers can include it
        return tokenStr, false, nil
    }

    // build new claims: copy existing claims except iat/exp
    newClaims := jwt.MapClaims{}
    for k, v := range claims {
        if k == "iat" || k == "exp" {
            continue
        }
        newClaims[k] = v
    }
    newClaims["iat"] = now.Unix()
    newClaims["exp"] = now.Add(tokenTtl).Unix()

    // choose signing method matching alg
    var signMethod jwt.SigningMethod
    switch alg {
    case "HS256":
        signMethod = jwt.SigningMethodHS256
    case "HS384":
        signMethod = jwt.SigningMethodHS384
    case "HS512":
        signMethod = jwt.SigningMethodHS512
    default:
        // fallback to HS512
        signMethod = jwt.SigningMethodHS512
    }

    newToken := jwt.NewWithClaims(signMethod, newClaims)
    signed, err := newToken.SignedString(secretBytes)
    if err != nil {
        return "", false, fmt.Errorf("failed to sign refreshed token: %w", err)
    }

    return signed, true, nil
}
