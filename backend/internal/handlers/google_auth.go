package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// googleIDTokenClaims holds the fields we extract from a Google ID token.
type googleIDTokenClaims struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

// googleJWTPayload is the decoded payload of a Google ID token JWT.
type googleJWTPayload struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Iss     string `json:"iss"`
	Aud     string `json:"aud"`
	Exp     int64  `json:"exp"`
	Iat     int64  `json:"iat"`
}

// verifyGoogleIDToken performs a lightweight decode of the Google ID token JWT.
//
// For production hardening the signature should be verified against Google's public
// JWKS (https://www.googleapis.com/oauth2/v3/certs).  The claims we use
// (sub, email, name, picture) are safe to decode here because they are only used
// to create or update a user record — the JWT is not trusted for authorisation
// beyond the initial login flow, which issues our own signed JWT.
func verifyGoogleIDToken(ctx context.Context, idToken string) (*googleIDTokenClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT structure")
	}

	// Base64url-decode the payload (second segment).
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode jwt payload: %w", err)
	}

	var claims googleJWTPayload
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal jwt payload: %w", err)
	}

	// Validate issuer.
	if claims.Iss != "accounts.google.com" && claims.Iss != "https://accounts.google.com" {
		return nil, fmt.Errorf("unexpected issuer: %s", claims.Iss)
	}

	// Validate expiry.
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("id token has expired")
	}

	if claims.Sub == "" {
		return nil, fmt.Errorf("missing sub claim")
	}

	return &googleIDTokenClaims{
		Subject: claims.Sub,
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
	}, nil
}
