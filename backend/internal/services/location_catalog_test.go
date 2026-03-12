package services

import (
	"encoding/json"
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
