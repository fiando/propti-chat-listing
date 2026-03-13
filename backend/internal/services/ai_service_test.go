package services

import (
	"encoding/json"
	"strings"
	"testing"

	openai "github.com/sashabaranov/go-openai"

	"github.com/fiando/propti/backend/internal/models"
)

func TestBuildParseChatCompletionRequestUsesFastParserModel(t *testing.T) {
	t.Parallel()

	req := buildParseChatCompletionRequest("jual rumah depok")

	if req.Model != "gpt-4o-mini" {
		t.Fatalf("expected fast parser model %q, got %q", "gpt-4o-mini", req.Model)
	}

	if req.Temperature != 0 {
		t.Fatalf("expected parse request temperature to remain unset/zero, got %v", req.Temperature)
	}

	if req.ResponseFormat == nil || req.ResponseFormat.Type != openai.ChatCompletionResponseFormatTypeJSONObject {
		t.Fatalf("expected JSON response format, got %#v", req.ResponseFormat)
	}
}

func TestParseListingTextSupportsLocationSuggestions(t *testing.T) {
	t.Parallel()

	payload := `{
		"title":"Rumah siap huni di Beji",
		"description":"Rumah bagus dekat kampus",
		"price":850000000,
		"priceUnit":"total",
		"propertyDetails":{"landArea":120,"buildingArea":90,"bedrooms":3,"bathrooms":2,"amenities":[]},
		"address":"Depok Beji dkt tol Cijago",
		"locationSuggestion":{
			"province":"Jawa Barat",
			"city":"Depok",
			"district":"Beji",
			"normalizedAddress":"Jalan Margonda Raya, Beji, Depok",
			"confidence":0.88
		},
		"confidence":0.92,
		"requiresManualReview":false,
		"warnings":[]
	}`

	var parsed models.ParsedListing
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		t.Fatalf("unmarshal parsed listing: %v", err)
	}

	if parsed.LocationSuggestion.City != "Depok" {
		t.Fatalf("expected suggested city Depok, got %q", parsed.LocationSuggestion.City)
	}

	if parsed.LocationSuggestion.Confidence != 0.88 {
		t.Fatalf("expected locationSuggestion.confidence 0.88, got %v", parsed.LocationSuggestion.Confidence)
	}
}

func TestParseSystemPromptIncludesFormattingGuidance(t *testing.T) {
	t.Parallel()

	expectedSnippets := []string{
		"Format \"title\" as clean headline-style Indonesian copy",
		"Format \"description\" as clean multiline Indonesian copy using short sections and emoji bullets when helpful.",
		"Use newline characters in the JSON string to separate lines.",
		"Format \"address\" as a clean, readable full address string for a form field",
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(parseSystemPrompt, snippet) {
			t.Fatalf("expected parse prompt to include %q", snippet)
		}
	}
}
