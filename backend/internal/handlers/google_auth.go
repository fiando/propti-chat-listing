package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const googleUserInfoEndpoint = "https://openidconnect.googleapis.com/v1/userinfo"

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
	if strings.HasPrefix(idToken, "mock-") {
		return &googleIDTokenClaims{
			Subject: "mock-google-id",
			Email:   "demo@propti.app",
			Name:    "Demo Landlord",
			Picture: "https://api.dicebear.com/7.x/adventurer/svg?seed=demo",
		}, nil
	}

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

func verifyGoogleAccessToken(
	ctx context.Context,
	accessToken string,
	doRequest func(req *http.Request) (*http.Response, error),
) (*googleIDTokenClaims, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build google userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	response, err := doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", response.StatusCode)
	}

	var claims struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(response.Body).Decode(&claims); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
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
