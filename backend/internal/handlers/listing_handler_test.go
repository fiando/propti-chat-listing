package handlers

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestParseSearchParamsSupportsComprehensiveFilters(t *testing.T) {
	req := events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{
			"q":               "rumah keluarga",
			"province":        "DKI Jakarta",
			"city":            "Jakarta Selatan",
			"listingType":     "sell",
			"priceMin":        "500000000",
			"priceMax":        "2000000000",
			"bedrooms":        "3",
			"bathrooms":       "2",
			"buildingAreaMin": "90",
			"buildingAreaMax": "180",
			"landAreaMin":     "120",
			"landAreaMax":     "250",
			"legalStatus":     "SHM - Sertifikat Hak Milik",
			"amenities":       "carport,kolam_renang",
			"sortBy":          "popular",
			"page":            "2",
			"pageSize":        "12",
		},
	}

	params := parseSearchParams(req)

	if params.Query != "rumah keluarga" {
		t.Fatalf("expected query to be parsed, got %q", params.Query)
	}
	if params.Bathrooms != 2 {
		t.Fatalf("expected bathrooms=2, got %d", params.Bathrooms)
	}
	if params.BuildingAreaMin != 90 || params.BuildingAreaMax != 180 {
		t.Fatalf("expected building area min/max 90-180, got %v-%v", params.BuildingAreaMin, params.BuildingAreaMax)
	}
	if params.LandAreaMin != 120 || params.LandAreaMax != 250 {
		t.Fatalf("expected land area min/max 120-250, got %v-%v", params.LandAreaMin, params.LandAreaMax)
	}
	if params.LegalStatus != "SHM - Sertifikat Hak Milik" {
		t.Fatalf("expected legal status to be parsed, got %q", params.LegalStatus)
	}
	if !reflect.DeepEqual(params.Amenities, []string{"carport", "kolam_renang"}) {
		t.Fatalf("unexpected amenities: %#v", params.Amenities)
	}
}

func TestTemplateExposesListingViewAndContactRevealRoutes(t *testing.T) {
	template, err := os.ReadFile("../../template.yaml")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	content := string(template)

	if !strings.Contains(content, "Path: /listings/{id}/view") {
		t.Fatal("expected SAM template to expose POST /listings/{id}/view")
	}
	if !strings.Contains(content, "Path: /listings/{id}/contact-reveal") {
		t.Fatal("expected SAM template to expose POST /listings/{id}/contact-reveal")
	}
}
