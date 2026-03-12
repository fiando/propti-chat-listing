package services

import (
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
