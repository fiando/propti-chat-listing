package services

import (
	"encoding/json"
	"testing"

	"github.com/fiando/propti/backend/internal/models"
)

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
			"normalizedAddress":"Jalan Margonda Raya, Beji, Depok"
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
}
