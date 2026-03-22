package services

import (
	"strings"
	"testing"

	"github.com/fiando/propti/backend/internal/models"
)

func TestNormalizeSearchIntentMapsCanonicalSearchFilters(t *testing.T) {
	t.Parallel()

	catalog := mustBuildLocationCatalog(t)
	service := NewSearchIntentService(nil, catalog)

	intent := &models.SearchIntent{
		Query:           "rumah dijual jogja harga termurah",
		KeywordQuery:    "rumah keluarga",
		ListingType:     string(models.ListingTypeSell),
		Province:        "jogja",
		City:            "gunung kidul",
		PriceMin:        500000000,
		PriceMax:        1000000000,
		Bedrooms:        1,
		Bathrooms:       3,
		BuildingAreaMin: 50,
		LandAreaMin:     100,
		LegalStatus:     "sertifikat hak milik",
		Amenities:       []string{"AC", "CCTV", "wifi"},
		SortBy:          "harga terendah",
	}

	params, metadata := service.Normalize(intent)

	if params.Query != "" {
		t.Fatalf("expected keyword query to be dropped when structured filters are already populated, got %q", params.Query)
	}
	if params.ListingType != models.ListingTypeSell {
		t.Fatalf("expected sell listing type, got %q", params.ListingType)
	}
	if params.Province != "DI Yogyakarta" {
		t.Fatalf("expected normalized province DI Yogyakarta, got %q", params.Province)
	}
	if params.City != "Gunungkidul" {
		t.Fatalf("expected normalized city Gunungkidul, got %q", params.City)
	}
	if params.LegalStatus != "SHM" {
		t.Fatalf("expected normalized legal status SHM, got %q", params.LegalStatus)
	}
	if got := strings.Join(params.Amenities, ","); got != "ac,cctv,internet_wifi" {
		t.Fatalf("expected normalized amenities, got %q", got)
	}
	if params.SortBy != "price_asc" {
		t.Fatalf("expected normalized sort price_asc, got %q", params.SortBy)
	}
	if !metadata.LocationResolved {
		t.Fatal("expected location to resolve against catalog")
	}
}

func TestNormalizeSearchIntentKeepsKeywordQueryWhenNoStructuredFiltersExist(t *testing.T) {
	t.Parallel()

	service := NewSearchIntentService(nil, nil)

	intent := &models.SearchIntent{
		Query:        "rumah dekat kampus UGM",
		KeywordQuery: "dekat kampus UGM",
	}

	params, _ := service.Normalize(intent)

	if params.Query != "dekat kampus UGM" {
		t.Fatalf("expected keyword query to be preserved for free-text searches, got %q", params.Query)
	}
}

func TestNormalizeSearchIntentDropsUnknownAmenitiesAndSort(t *testing.T) {
	t.Parallel()

	service := NewSearchIntentService(nil, nil)

	intent := &models.SearchIntent{
		Query:      "rumah murah",
		Amenities:  []string{"helipad", "cctv"},
		LegalStatus:"girik",
		SortBy:     "termurah banget",
	}

	params, metadata := service.Normalize(intent)

	if got := strings.Join(params.Amenities, ","); got != "cctv" {
		t.Fatalf("expected only known amenities to remain, got %q", got)
	}
	if params.SortBy != "" {
		t.Fatalf("expected unsupported sort to be cleared, got %q", params.SortBy)
	}
	if params.LegalStatus != "Girik" {
		t.Fatalf("expected legal status Girik, got %q", params.LegalStatus)
	}
	if metadata.LocationResolved {
		t.Fatal("expected no location resolution without location input")
	}
}

func TestNormalizeSearchIntentSnapsSingleBudgetToPresetRange(t *testing.T) {
	t.Parallel()

	service := NewSearchIntentService(nil, nil)

	intent := &models.SearchIntent{
		Query:        "rumah dijual harga kisaran 500 juta",
		ListingType:  string(models.ListingTypeSell),
		PriceMin:     500000000,
		PriceMax:     500000000,
	}

	params, _ := service.Normalize(intent)

	if params.PriceMin != 500000000 || params.PriceMax != 1000000000 {
		t.Fatalf("expected single-point 500 juta budget to snap to 500 juta - 1 miliar preset, got %v-%v", params.PriceMin, params.PriceMax)
	}
}

func TestNormalizeSearchIntentDropsKeywordQueryWhenStructuredFiltersAlreadyCoverIntent(t *testing.T) {
	t.Parallel()

	service := NewSearchIntentService(nil, nil)

	intent := &models.SearchIntent{
		Query:        "rumah dijual yogyakarta KT 4 KM 2 SHM ada ruang tamu carport dapur harga kisaran 500 juta",
		KeywordQuery: "yogyakarta 4 2 SHM ruang tamu carport dapur 500 juta",
		ListingType:  string(models.ListingTypeSell),
		Province:     "DI Yogyakarta",
		PriceMin:     500000000,
		PriceMax:     500000000,
		Bedrooms:     4,
		Bathrooms:    2,
		LegalStatus:  "SHM",
		Amenities:    []string{"ruang_tamu", "carport", "dapur"},
	}

	params, _ := service.Normalize(intent)

	if params.Query != "" {
		t.Fatalf("expected keyword query to be cleared when structured filters already capture the intent, got %q", params.Query)
	}
}

func mustBuildLocationCatalog(t *testing.T) *LocationCatalog {
	t.Helper()

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(`{
		"provinces":[{"id":"34","name":"DI Yogyakarta"}],
		"cities":[{"id":"3403","provinceId":"34","name":"Gunungkidul"}],
		"districts":[]
	}`))
	if err != nil {
		t.Fatalf("build location catalog: %v", err)
	}
	return catalog
}
