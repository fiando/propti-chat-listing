package handlers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestVerifyGoogleAccessTokenFetchesUserInfoClaims(t *testing.T) {
	claims, err := verifyGoogleAccessToken(context.Background(), "google-access-token", func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://openidconnect.googleapis.com/v1/userinfo" {
			t.Fatalf("expected userinfo endpoint, got %s", req.URL.String())
		}
		if got := req.Header.Get("Authorization"); got != "Bearer google-access-token" {
			t.Fatalf("expected bearer access token, got %q", got)
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"sub":"google-user-123",
				"email":"user@example.com",
				"name":"Propti User",
				"picture":"https://example.com/avatar.png"
			}`)),
		}, nil
	})
	if err != nil {
		t.Fatalf("verifyGoogleAccessToken returned error: %v", err)
	}

	if claims.Subject != "google-user-123" {
		t.Fatalf("expected subject google-user-123, got %q", claims.Subject)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %q", claims.Email)
	}
	if claims.Name != "Propti User" {
		t.Fatalf("expected name Propti User, got %q", claims.Name)
	}
	if claims.Picture != "https://example.com/avatar.png" {
		t.Fatalf("expected picture URL to be returned, got %q", claims.Picture)
	}
}
