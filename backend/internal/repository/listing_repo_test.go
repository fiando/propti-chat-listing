package repository

import (
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/fiando/propti/backend/internal/models"
)

func TestBuildListingScanQueryTargetsNestedBedroomsField(t *testing.T) {
	query := buildListingScanQuery(&models.ListingSearchParams{
		Bedrooms: 3,
	})

	if !strings.Contains(query.filterExpression, "#details.#bedrooms >= :bed") {
		t.Fatalf("expected nested bedroom filter, got %q", query.filterExpression)
	}
	if _, ok := query.expressionAttributeNames["#details"]; !ok {
		t.Fatalf("expected propertyDetails attribute name to be registered")
	}
	if _, ok := query.expressionAttributeNames["#bedrooms"]; !ok {
		t.Fatalf("expected bedrooms attribute name to be registered")
	}
	if got, ok := query.expressionAttributeValues[":bed"].(*types.AttributeValueMemberN); !ok || got.Value != "3" {
		t.Fatalf("expected numeric bedroom filter value 3, got %#v", query.expressionAttributeValues[":bed"])
	}
}

func TestBuildListingScanQuerySupportsComprehensivePropertyFilters(t *testing.T) {
	query := buildListingScanQuery(&models.ListingSearchParams{
		Query:           "cluster jakarta",
		Bathrooms:       2,
		BuildingAreaMin: 80,
		BuildingAreaMax: 150,
		LandAreaMin:     100,
		LandAreaMax:     220,
		LegalStatus:     "SHM - Sertifikat Hak Milik",
		Amenities:       []string{"carport", "kolam_renang"},
	})

	expectedFragments := []string{
		"(contains(#title, :query)",
		"contains(#description, :query)",
		"contains(#loc.#address, :query)",
		"#details.#bathrooms >= :bath",
		"#details.#buildingArea >= :buildingAreaMin",
		"#details.#buildingArea <= :buildingAreaMax",
		"#details.#landArea >= :landAreaMin",
		"#details.#landArea <= :landAreaMax",
		"#details.#legalStatus = :legalStatus",
		"contains(#details.#amenities, :amenity0)",
		"contains(#details.#amenities, :amenity1)",
	}

	for _, fragment := range expectedFragments {
		if !strings.Contains(query.filterExpression, fragment) {
			t.Fatalf("expected filter expression to contain %q, got %q", fragment, query.filterExpression)
		}
	}

	if got, ok := query.expressionAttributeValues[":bath"].(*types.AttributeValueMemberN); !ok || got.Value != "2" {
		t.Fatalf("expected bathroom filter value 2, got %#v", query.expressionAttributeValues[":bath"])
	}
	if got, ok := query.expressionAttributeValues[":legalStatus"].(*types.AttributeValueMemberS); !ok || got.Value != "SHM - Sertifikat Hak Milik" {
		t.Fatalf("expected legalStatus filter value, got %#v", query.expressionAttributeValues[":legalStatus"])
	}
	if got, ok := query.expressionAttributeValues[":amenity0"].(*types.AttributeValueMemberS); !ok || got.Value != "carport" {
		t.Fatalf("expected first amenity filter value, got %#v", query.expressionAttributeValues[":amenity0"])
	}
	if got, ok := query.expressionAttributeValues[":amenity1"].(*types.AttributeValueMemberS); !ok || got.Value != "kolam_renang" {
		t.Fatalf("expected second amenity filter value, got %#v", query.expressionAttributeValues[":amenity1"])
	}
}

func TestSortListingsAppliesRequestedOrdering(t *testing.T) {
	listings := []models.Listing{
		{ListingID: "mid-price", Price: 700000000, Views: 10, CreatedAt: mustParseTime(t, "2026-03-11T00:00:00Z")},
		{ListingID: "highest-price", Price: 1200000000, Views: 2, CreatedAt: mustParseTime(t, "2026-03-10T00:00:00Z")},
		{ListingID: "lowest-price", Price: 500000000, Views: 20, CreatedAt: mustParseTime(t, "2026-03-12T00:00:00Z")},
	}

	priceDesc := append([]models.Listing(nil), listings...)
	sortListings(priceDesc, "price_desc")
	if got := []string{priceDesc[0].ListingID, priceDesc[1].ListingID, priceDesc[2].ListingID}; strings.Join(got, ",") != "highest-price,mid-price,lowest-price" {
		t.Fatalf("unexpected price_desc order: %v", got)
	}

	popular := append([]models.Listing(nil), listings...)
	sortListings(popular, "popular")
	if got := []string{popular[0].ListingID, popular[1].ListingID, popular[2].ListingID}; strings.Join(got, ",") != "lowest-price,mid-price,highest-price" {
		t.Fatalf("unexpected popular order: %v", got)
	}
}

func mustParseTime(t *testing.T, value string) time.Time {
	t.Helper()

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed
}
