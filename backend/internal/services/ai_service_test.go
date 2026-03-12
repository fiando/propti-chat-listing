package services

import (
	"encoding/json"
	"testing"

	openai "github.com/sashabaranov/go-openai"

	"github.com/fiando/propti/backend/internal/models"
)

func TestBuildParseChatCompletionRequestUsesDefaultTemperatureForGPT5(t *testing.T) {
	t.Parallel()

	req := buildParseChatCompletionRequest("jual rumah depok")

	if req.Model != parserModel {
		t.Fatalf("expected parser model %q, got %q", parserModel, req.Model)
	}

	if req.Temperature != 0 {
		t.Fatalf("expected parse request to omit temperature for GPT-5, got %v", req.Temperature)
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
