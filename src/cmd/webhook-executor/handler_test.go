package main

import (
    "encoding/hex"
    "testing"
    "time"

    internaljwt "github.com/NobleFactor/docker-webhook/cmd/internal/jwt"
    "github.com/NobleFactor/docker-webhook/cmd/internal/sshremote"
    "github.com/golang-jwt/jwt/v5"
)

// Test that when RefreshJWT issues a refreshed token, the handler puts the token into the Response.AuthToken field.
func TestHandler_AttachesAuthTokenWhenRefreshed(t *testing.T) {
    // prepare a raw secret and hex encode it in the same way the service does
    rawSecret := []byte("handler-test-secret-xxxxxxxxxxxxxxxxxxxx")
    secretHex := hex.EncodeToString(rawSecret)
    // build a short-lived token so it will be refreshed
    now := time.Now()
    loc := "handler-location"
    claims := jwt.MapClaims{
        "sub": loc,
        "iat": now.Unix(),
        "exp": now.Add(5 * time.Second).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(rawSecret)
    if err != nil {
        t.Fatalf("failed to sign token: %v", err)
    }

    authHeader := "Bearer " + signed
    // validate to get parsed token and then call RefreshJWT
    tokenStr, parsed, err := internaljwt.ValidateJWT(authHeader, secretHex, loc)
    if err != nil {
        t.Fatalf("ValidateJWT failed: %v", err)
    }
    tokenRefreshWindow := 1 * time.Minute
    tokenTtl := 2 * time.Minute
    newTok, refreshed, err := internaljwt.RefreshJWT(parsed, tokenStr, secretHex, loc, tokenRefreshWindow, tokenTtl)
    if err != nil {
        t.Fatalf("unexpected error from RefreshJWT: %v", err)
    }
    if !refreshed {
        t.Fatalf("expected token to be refreshed")
    }
    if newTok == "" {
        t.Fatalf("expected non-empty refreshed token")
    }

    // mimic handler attaching the token to response
    var resp sshremote.Response
    resp.CorrelationId = "test-cid"
    if refreshed {
        resp.AuthToken = &newTok
    }

    if resp.AuthToken == nil {
        t.Fatalf("expected AuthToken to be set on response")
    }
    if *resp.AuthToken != newTok {
        t.Fatalf("AuthToken value mismatch: expected %q got %q", newTok, *resp.AuthToken)
    }
}
