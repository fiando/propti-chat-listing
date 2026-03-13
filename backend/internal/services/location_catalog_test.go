package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestLocationCatalogFindsCitiesAndDistricts(t *testing.T) {
	t.Parallel()

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(`{
		"provinces":[{"id":"32","name":"Jawa Barat"}],
		"cities":[{"id":"3276","provinceId":"32","name":"Depok"}],
		"districts":[{"id":"3276010","cityId":"3276","name":"Beji"}]
	}`))
	if err != nil {
		t.Fatalf("NewLocationCatalogFromReader returned error: %v", err)
	}

	cities := catalog.SearchCities("32", "dep")
	if len(cities) != 1 || cities[0].Name != "Depok" {
		t.Fatalf("expected Depok city result, got %#v", cities)
	}

	districts := catalog.SearchDistricts("3276", "bej")
	if len(districts) != 1 || districts[0].Name != "Beji" {
		t.Fatalf("expected Beji district result, got %#v", districts)
	}
}

func TestLocationCatalogSearchProvinces(t *testing.T) {
	t.Parallel()

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(`{
		"provinces":[
			{"id":"32","name":"Jawa Barat"},
			{"id":"33","name":"Jawa Tengah"},
			{"id":"51","name":"Bali"}
		],
		"cities":[],
		"districts":[]
	}`))
	if err != nil {
		t.Fatalf("NewLocationCatalogFromReader returned error: %v", err)
	}

	results := catalog.SearchProvinces("jawa")
	if len(results) != 2 {
		t.Fatalf("expected 2 provinces matching 'jawa', got %d: %#v", len(results), results)
	}

	results = catalog.SearchProvinces("BALI")
	if len(results) != 1 || results[0].Name != "Bali" {
		t.Fatalf("expected Bali province result, got %#v", results)
	}

	results = catalog.SearchProvinces("xyz")
	if len(results) != 0 {
		t.Fatalf("expected no results for 'xyz', got %#v", results)
	}
}

func TestLocationCatalogSearchCapsResultWindows(t *testing.T) {
	t.Parallel()

	var provinceEntries []string
	var cityEntries []string
	var districtEntries []string
	for i := 1; i <= 25; i++ {
		provinceEntries = append(provinceEntries, fmt.Sprintf(`{"id":"p%02d","name":"Provinsi %02d"}`, i, i))
		cityEntries = append(cityEntries, fmt.Sprintf(`{"id":"c%02d","provinceId":"p01","name":"Kota %02d"}`, i, i))
		districtEntries = append(districtEntries, fmt.Sprintf(`{"id":"d%02d","cityId":"c01","name":"Kecamatan %02d"}`, i, i))
	}

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(fmt.Sprintf(`{
		"provinces":[%s],
		"cities":[%s],
		"districts":[%s]
	}`, strings.Join(provinceEntries, ","), strings.Join(cityEntries, ","), strings.Join(districtEntries, ","))))
	if err != nil {
		t.Fatalf("NewLocationCatalogFromReader returned error: %v", err)
	}

	// Non-empty query: cap at maxSearchResults (20).
	provinces := catalog.SearchProvinces("provinsi")
	if len(provinces) != 20 {
		t.Fatalf("expected capped province results of 20, got %d", len(provinces))
	}

	cities := catalog.SearchCities("p01", "kota")
	if len(cities) != 20 {
		t.Fatalf("expected capped city results of 20, got %d", len(cities))
	}

	districts := catalog.SearchDistricts("c01", "kecamatan")
	if len(districts) != 20 {
		t.Fatalf("expected capped district results of 20, got %d", len(districts))
	}

	// Empty query: all items must be returned (dropdown population use-case).
	allProvinces := catalog.SearchProvinces("")
	if len(allProvinces) != 25 {
		t.Fatalf("expected all 25 provinces for empty query, got %d", len(allProvinces))
	}

	allCities := catalog.SearchCities("p01", "")
	if len(allCities) != 25 {
		t.Fatalf("expected all 25 cities for empty query, got %d", len(allCities))
	}

	allDistricts := catalog.SearchDistricts("c01", "")
	if len(allDistricts) != 25 {
		t.Fatalf("expected all 25 districts for empty query, got %d", len(allDistricts))
	}
}

func TestLocationCatalogEmptyResultsSerializeAsArray(t *testing.T) {
	t.Parallel()

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(`{
"provinces":[{"id":"32","name":"Jawa Barat"}],
"cities":[{"id":"3276","provinceId":"32","name":"Depok"}],
"districts":[{"id":"3276010","cityId":"3276","name":"Beji"}]
}`))
	if err != nil {
		t.Fatalf("NewLocationCatalogFromReader returned error: %v", err)
	}

	provinces := catalog.SearchProvinces("nomatch")
	cities := catalog.SearchCities("32", "nomatch")
	districts := catalog.SearchDistricts("3276", "nomatch")

	for name, v := range map[string]any{
		"provinces": provinces,
		"cities":    cities,
		"districts": districts,
	} {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("json.Marshal(%s) error: %v", name, err)
		}
		if string(b) != "[]" {
			t.Errorf("%s: expected JSON `[]`, got `%s`", name, b)
		}
	}
}

func TestNormalizeSuggestionExactPreferredOverSubstring(t *testing.T) {
	t.Parallel()

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(`{
"provinces":[
{"id":"51","name":"Bali"},
{"id":"52","name":"Nusa Tenggara Barat"}
],
"cities":[],
"districts":[]
}`))
	if err != nil {
		t.Fatalf("NewLocationCatalogFromReader returned error: %v", err)
	}

	result := catalog.NormalizeSuggestion("Bali", "", "")
	if result.Province != "Bali" {
		t.Fatalf("expected exact match 'Bali', got %q", result.Province)
	}
	if result.Confidence < 0.99 {
		t.Fatalf("expected confidence=1.0 for exact match, got %v", result.Confidence)
	}
}

func TestNormalizeSuggestionShortFragmentNotMatched(t *testing.T) {
	t.Parallel()

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(`{
"provinces":[{"id":"32","name":"Jawa Barat"}],
"cities":[{"id":"3276","provinceId":"32","name":"Depok"}],
"districts":[]
}`))
	if err != nil {
		t.Fatalf("NewLocationCatalogFromReader returned error: %v", err)
	}

	// "De" is too short for substring matching and must not match "Depok".
	result := catalog.NormalizeSuggestion("Jawa Barat", "De", "")
	if result.City == "Depok" {
		t.Fatalf("short fragment 'De' must not match 'Depok' via substring, got city %q", result.City)
	}
	// Only province contributed to confidence.
	if result.Confidence > 0.51 {
		t.Fatalf("expected confidence ~0.5 (province only), got %v", result.Confidence)
	}
}

// TestNormalizeSuggestionWrongProvinceRightCity is the regression test for the
// broad city fallback: when the AI supplies a province that exists in the
// catalog but the requested city belongs to a *different* province, the
// suggestion must carry the city's actual province (not the AI's wrong one) and
// the confidence must not be inflated by counting the wrong province as a hit.
func TestNormalizeSuggestionWrongProvinceRightCity(t *testing.T) {
	t.Parallel()

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(`{
"provinces":[
	{"id":"32","name":"Jawa Barat"},
	{"id":"33","name":"Jawa Tengah"}
],
"cities":[
	{"id":"3276","provinceId":"32","name":"Depok"},
	{"id":"3374","provinceId":"33","name":"Semarang"}
],
"districts":[]
}`))
	if err != nil {
		t.Fatalf("NewLocationCatalogFromReader returned error: %v", err)
	}

	// AI gave Jawa Tengah as province, but Depok is in Jawa Barat.
	result := catalog.NormalizeSuggestion("Jawa Tengah", "Depok", "")
	if result.Province != "Jawa Barat" {
		t.Fatalf("expected province corrected to 'Jawa Barat', got %q", result.Province)
	}
	if result.City != "Depok" {
		t.Fatalf("expected city='Depok', got %q", result.City)
	}
	// Province was wrong; only city resolved → confidence = 1/2.
	const want = 1.0 / 2.0
	if result.Confidence < want-0.01 || result.Confidence > want+0.01 {
		t.Fatalf("expected confidence ~%.4f, got %v", want, result.Confidence)
	}
}

func TestNormalizeSuggestionCityResolvesWhenProvinceUnmatched(t *testing.T) {
	t.Parallel()

	catalog, err := NewLocationCatalogFromReader(strings.NewReader(`{
"provinces":[{"id":"32","name":"Jawa Barat"}],
"cities":[{"id":"3276","provinceId":"32","name":"Depok"}],
"districts":[{"id":"3276010","cityId":"3276","name":"Beji"}]
}`))
	if err != nil {
		t.Fatalf("NewLocationCatalogFromReader returned error: %v", err)
	}

	// Province is wrong, but city and district are valid.
	result := catalog.NormalizeSuggestion("Unknown Province", "Depok", "Beji")
	if result.City != "Depok" {
		t.Fatalf("expected city='Depok' via broad fallback, got %q", result.City)
	}
	if result.District != "Beji" {
		t.Fatalf("expected district='Beji', got %q", result.District)
	}
	// Province inferred from city should update the display field.
	if result.Province != "Jawa Barat" {
		t.Fatalf("expected province inferred as 'Jawa Barat', got %q", result.Province)
	}
	// Province input didn't resolve directly: confidence = city + district = 2/3.
	const want = 2.0 / 3.0
	if result.Confidence < want-0.01 || result.Confidence > want+0.01 {
		t.Fatalf("expected confidence ~%.4f, got %v", want, result.Confidence)
	}
}
