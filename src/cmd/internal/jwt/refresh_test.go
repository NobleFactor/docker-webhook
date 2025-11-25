package jwt

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// helper to build an HMAC-signed token with given alg and claims
func buildHMACToken(t *testing.T, alg string, secret []byte, claims jwt.MapClaims) string {
	t.Helper()
	var method jwt.SigningMethod
	switch alg {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "HS384":
		method = jwt.SigningMethodHS384
	case "HS512":
		method = jwt.SigningMethodHS512
	default:
		method = jwt.SigningMethodHS512
	}

	token := jwt.NewWithClaims(method, claims)
	s, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return s
}

func TestRefreshJWT_NoRefresh(t *testing.T) {
	secret := []byte("test-secret-01234567890123456789")
	secretHex := hex.EncodeToString(secret)

	now := time.Now()
	// set exp well beyond refresh window
	claims := jwt.MapClaims{
		"sub": "location-a",
		"iat": now.Unix(),
		"exp": now.Add(10 * time.Minute).Unix(),
	}

	tok := buildHMACToken(t, "HS512", secret, claims)
	tokenStr, parsed, err := ValidateJWT("Bearer "+tok, secretHex, "location-a")
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	newTok, refreshed, err := RefreshJWT(parsed, tokenStr, secretHex, "location-a", 1*time.Minute, 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refreshed {
		t.Fatalf("expected no refresh but got refreshed token: %q", newTok)
	}
	if newTok != tokenStr {
		t.Fatalf("expected returned token to equal original token when not refreshed; got %q", newTok)
	}
}

func TestRefreshJWT_Refresh(t *testing.T) {
	secret := []byte("another-test-secret-xxxxxxxxxxxxxxxx")
	secretHex := hex.EncodeToString(secret)

	now := time.Now()
	// set exp small so it triggers refresh (within window)
	claims := jwt.MapClaims{
		"sub":    "loc-1",
		"iat":    now.Unix(),
		"exp":    now.Add(10 * time.Second).Unix(),
		"custom": "value",
	}

	origAlg := "HS256"
	tok := buildHMACToken(t, origAlg, secret, claims)

	tokenStr, parsed, err := ValidateJWT("Bearer "+tok, secretHex, "loc-1")
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	newTok, refreshed, err := RefreshJWT(parsed, tokenStr, secretHex, "loc-1", 1*time.Minute, 2*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !refreshed {
		t.Fatalf("expected refreshed=true")
	}
	if newTok == "" {
		t.Fatalf("expected non-empty refreshed token")
	}

	// parse the refreshed token and validate signature and claims
	parsed2, err := jwt.Parse(newTok, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != origAlg {
			t.Fatalf("expected alg %s on refreshed token, got %s", origAlg, token.Method.Alg())
		}
		return secret, nil
	})
	if err != nil {
		t.Fatalf("failed to parse refreshed token: %v", err)
	}
	mclaims, ok := parsed2.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("refreshed token has unexpected claims type")
	}
	if mclaims["custom"] != "value" {
		t.Fatalf("expected custom claim preserved, got %#v", mclaims["custom"])
	}
	if expRaw, ok := mclaims["exp"].(float64); ok {
		expTime := time.Unix(int64(expRaw), 0)
		if expTime.Before(time.Now().Add(110 * time.Second)) {
			t.Fatalf("refreshed token exp too soon: %v", expTime)
		}
	} else {
		t.Fatalf("refreshed token missing exp claim or wrong type")
	}
}

func TestValidateJWT_InvalidSignature(t *testing.T) {
	secret := []byte("my-secret-1")
	otherSecret := []byte("my-secret-2")
	otherHex := hex.EncodeToString(otherSecret)

	now := time.Now()
	claims := jwt.MapClaims{
		"sub": "x",
		"iat": now.Unix(),
		"exp": now.Add(1 * time.Minute).Unix(),
	}

	tok := buildHMACToken(t, "HS512", secret, claims)

	_, _, err := ValidateJWT("Bearer "+tok, otherHex, "")
	if err == nil {
		t.Fatalf("expected ValidateJWT to fail with wrong secret, but it succeeded")
	}
}

func TestRefreshJWT_NonHMACAlg(t *testing.T) {
	headerJSON := `{"alg":"RS256","typ":"JWT"}`
	payloadJSON := `{"sub":"x","iat":0}`
	h := base64.RawURLEncoding.EncodeToString([]byte(headerJSON))
	p := base64.RawURLEncoding.EncodeToString([]byte(payloadJSON))
	tokenStr := h + "." + p + ".signature"

	// ValidateJWT should reject non-HMAC alg
	_, _, err := ValidateJWT("Bearer "+tokenStr, hex.EncodeToString([]byte("irrelevant")), "")
	if err == nil {
		t.Fatalf("expected ValidateJWT to error for non-HMAC alg token, got nil")
	}
}

func TestRefreshJWT_NoExp(t *testing.T) {
	secret := []byte("no-exp-secret-xxxxxxxxxxxxxxxxxxxx")
	secretHex := hex.EncodeToString(secret)

	now := time.Now()
	claims := jwt.MapClaims{
		"sub": "noexp",
		"iat": now.Unix(),
		"foo": "bar",
	}

	tok := buildHMACToken(t, "HS512", secret, claims)
	tokenStr, parsed, err := ValidateJWT("Bearer "+tok, secretHex, "noexp")
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}
	newTok, refreshed, err := RefreshJWT(parsed, tokenStr, secretHex, "noexp", 1*time.Minute, 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error when refreshing token without exp: %v", err)
	}
	if !refreshed {
		t.Fatalf("expected refresh for token without exp")
	}
	if newTok == "" {
		t.Fatalf("expected non-empty refreshed token")
	}
}
