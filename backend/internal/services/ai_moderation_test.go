package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIServiceModerateImagesAllowsHarmlessImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/moderations" {
			t.Fatalf("expected /v1/moderations path, got %s", r.URL.Path)
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode moderation request: %v", err)
		}

		if payload["model"] != openAIModerationModel {
			t.Fatalf("expected model %q, got %#v", openAIModerationModel, payload["model"])
		}

		input, ok := payload["input"].([]any)
		if !ok || len(input) != 1 {
			t.Fatalf("expected single multimodal input item, got %#v", payload["input"])
		}

		item, ok := input[0].(map[string]any)
		if !ok {
			t.Fatalf("expected multimodal input object, got %#v", input[0])
		}
		if item["type"] != "image_url" {
			t.Fatalf("expected image_url type, got %#v", item["type"])
		}

		imageURL, ok := item["image_url"].(map[string]any)
		if !ok {
			t.Fatalf("expected image_url payload, got %#v", item["image_url"])
		}
		url, _ := imageURL["url"].(string)
		if !strings.HasPrefix(url, "data:image/jpeg;base64,") {
			t.Fatalf("expected data URL image payload, got %q", url)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"mod-1","model":"omni-moderation-latest","results":[{"flagged":false,"categories":{"violence":false}}]}`))
	}))
	defer server.Close()

	service := &AIService{
		moderationAPIKey:  "test-key",
		moderationBaseURL: server.URL,
		httpClient:        server.Client(),
	}

	approved, reason, flags, err := service.ModerateImages(context.Background(), [][]byte{[]byte("image-bytes")})
	if err != nil {
		t.Fatalf("ModerateImages returned error: %v", err)
	}
	if !approved {
		t.Fatalf("expected harmless image to pass, got rejected with reason %q and flags %#v", reason, flags)
	}
	if reason != "" {
		t.Fatalf("expected empty reason, got %q", reason)
	}
	if len(flags) != 0 {
		t.Fatalf("expected no flags, got %#v", flags)
	}
}

func TestAIServiceModerateImagesRejectsFlaggedImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"mod-1","model":"omni-moderation-latest","results":[{"flagged":true,"categories":{"sexual":true,"violence":false}}]}`))
	}))
	defer server.Close()

	service := &AIService{
		moderationAPIKey:  "test-key",
		moderationBaseURL: server.URL,
		httpClient:        server.Client(),
	}

	approved, reason, flags, err := service.ModerateImages(context.Background(), [][]byte{[]byte("image-bytes")})
	if err != nil {
		t.Fatalf("ModerateImages returned error: %v", err)
	}
	if approved {
		t.Fatal("expected flagged image to be rejected")
	}
	if !strings.Contains(reason, "tidak pantas") {
		t.Fatalf("expected harmful-content rejection reason, got %q", reason)
	}
	if len(flags) == 0 || flags[0] != "sexual" {
		t.Fatalf("expected sexual flag, got %#v", flags)
	}
}

func TestAIServiceModerateContentRejectsHarmfulContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode moderation request: %v", err)
		}
		if payload["input"] == "" {
			t.Fatal("expected moderation input to include listing text")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"mod-1","model":"omni-moderation-latest","results":[{"flagged":true,"categories":{"harassment":true}}]}`))
	}))
	defer server.Close()

	service := &AIService{
		moderationAPIKey:  "test-key",
		moderationBaseURL: server.URL,
		httpClient:        server.Client(),
	}

	approved, reason, flags, err := service.ModerateContent(context.Background(), "Rumah murah", "Saya akan membunuh kamu")
	if err != nil {
		t.Fatalf("ModerateContent returned error: %v", err)
	}
	if approved {
		t.Fatal("expected harmful text to be rejected")
	}
	if !strings.Contains(reason, "tidak aman") {
		t.Fatalf("expected harmful-text rejection reason, got %q", reason)
	}
	if len(flags) == 0 || flags[0] != "harassment" {
		t.Fatalf("expected harassment flag, got %#v", flags)
	}
}

func TestAIServiceModerateContentAllowsHarmlessTextUsingOmniModerationOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"mod-1","model":"omni-moderation-latest","results":[{"flagged":false,"categories":{"violence":false}}]}`))
	}))
	defer server.Close()

	service := &AIService{
		moderationAPIKey:  "test-key",
		moderationBaseURL: server.URL,
		httpClient:        server.Client(),
	}

	approved, reason, flags, err := service.ModerateContent(context.Background(), "Rumah siap huni", "Dekat sekolah dan stasiun")
	if err != nil {
		t.Fatalf("ModerateContent returned error: %v", err)
	}
	if !approved {
		t.Fatalf("expected harmless text to pass, got rejected with reason %q and flags %#v", reason, flags)
	}
	if reason != "" {
		t.Fatalf("expected empty reason, got %q", reason)
	}
	if len(flags) != 0 {
		t.Fatalf("expected no flags, got %#v", flags)
	}
}
