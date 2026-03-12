package services

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/fiando/propti/backend/internal/models"
)

type Province struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type City struct {
	ID         string `json:"id"`
	ProvinceID string `json:"provinceId"`
	Name       string `json:"name"`
}

type District struct {
	ID     string `json:"id"`
	CityID string `json:"cityId"`
	Name   string `json:"name"`
}

type locationData struct {
	Provinces []Province `json:"provinces"`
	Cities    []City     `json:"cities"`
	Districts []District `json:"districts"`
}

type LocationCatalog struct {
	provinces        []Province
	citiesByProvince map[string][]City
	districtsByCity  map[string][]District
}

func NewLocationCatalogFromReader(r io.Reader) (*LocationCatalog, error) {
	var raw locationData
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}

	citiesByProvince := make(map[string][]City, len(raw.Provinces))
	for _, c := range raw.Cities {
		citiesByProvince[c.ProvinceID] = append(citiesByProvince[c.ProvinceID], c)
	}

	districtsByCity := make(map[string][]District, len(raw.Cities))
	for _, d := range raw.Districts {
		districtsByCity[d.CityID] = append(districtsByCity[d.CityID], d)
	}

	return &LocationCatalog{
		provinces:        raw.Provinces,
		citiesByProvince: citiesByProvince,
		districtsByCity:  districtsByCity,
	}, nil
}

// minSubstringLen is the minimum query length required for substring-based
// matching. Shorter fragments risk false positives (e.g. "De" → "Depok").
const minSubstringLen = 3

// findCityInList returns the first city whose name matches query using an
// exact case-insensitive comparison, falling back to substring search only
// when the query meets the minimum specificity threshold.
func findCityInList(cities []City, query string) *City {
	for i := range cities {
		if strings.EqualFold(cities[i].Name, query) {
			return &cities[i]
		}
	}
	if len(query) >= minSubstringLen {
		for i := range cities {
			if strings.Contains(strings.ToLower(cities[i].Name), strings.ToLower(query)) {
				return &cities[i]
			}
		}
	}
	return nil
}

func (c *LocationCatalog) SearchProvinces(query string) []Province {
	q := strings.ToLower(query)
	results := []Province{}
	for _, p := range c.provinces {
		if strings.Contains(strings.ToLower(p.Name), q) {
			results = append(results, p)
		}
	}
	return results
}

func (c *LocationCatalog) SearchCities(provinceID, query string) []City {
	q := strings.ToLower(query)
	results := []City{}
	for _, city := range c.citiesByProvince[provinceID] {
		if strings.Contains(strings.ToLower(city.Name), q) {
			results = append(results, city)
		}
	}
	return results
}

func (c *LocationCatalog) SearchDistricts(cityID, query string) []District {
	q := strings.ToLower(query)
	results := []District{}
	for _, d := range c.districtsByCity[cityID] {
		if strings.Contains(strings.ToLower(d.Name), q) {
			results = append(results, d)
		}
	}
	return results
}

func (c *LocationCatalog) Provinces() []Province {
	return c.provinces
}

// NormalizeSuggestion validates the AI-suggested province/city/district names
// against the catalog and returns a scored ParsedLocationSuggestion.
// Confidence is computed as the fraction of supplied, non-empty fields that
// resolve to an authoritative catalog entry (each field contributes 1/3).
//
// If province does not match, city resolution is still attempted across all
// provinces so that a single bad province input does not zero out the entire
// confidence score.
func (c *LocationCatalog) NormalizeSuggestion(province, city, district string) models.ParsedLocationSuggestion {
	suggestion := models.ParsedLocationSuggestion{
		Province: province,
		City:     city,
		District: district,
	}

	// --- province resolution (exact first, substring fallback) ---
	var matchedProvince *Province
	if province != "" {
		for i := range c.provinces {
			if strings.EqualFold(c.provinces[i].Name, province) {
				matchedProvince = &c.provinces[i]
				suggestion.Province = c.provinces[i].Name
				break
			}
		}
		if matchedProvince == nil && len(province) >= minSubstringLen {
			for i := range c.provinces {
				if strings.Contains(strings.ToLower(c.provinces[i].Name), strings.ToLower(province)) {
					matchedProvince = &c.provinces[i]
					suggestion.Province = c.provinces[i].Name
					break
				}
			}
		}
	}

	// --- city resolution ---
	// Try within the matched province first; if still unresolved, broaden to
	// all provinces so a wrong/missing province doesn't block city resolution.
	var matchedCity *City
	if city != "" {
		if matchedProvince != nil {
			matchedCity = findCityInList(c.citiesByProvince[matchedProvince.ID], city)
		}
		if matchedCity == nil {
			for _, cities := range c.citiesByProvince {
				if ct := findCityInList(cities, city); ct != nil {
					matchedCity = ct
					// When province was not directly matched, infer it from the city.
					if matchedProvince == nil {
						for i := range c.provinces {
							if c.provinces[i].ID == matchedCity.ProvinceID {
								suggestion.Province = c.provinces[i].Name
								break
							}
						}
					}
					break
				}
			}
		}
		if matchedCity != nil {
			suggestion.City = matchedCity.Name
		}
	}

	// --- district resolution (exact first, substring fallback) ---
	var matchedDistrict bool
	if district != "" && matchedCity != nil {
		districts := c.districtsByCity[matchedCity.ID]
		for _, d := range districts {
			if strings.EqualFold(d.Name, district) {
				suggestion.District = d.Name
				matchedDistrict = true
				break
			}
		}
		if !matchedDistrict && len(district) >= minSubstringLen {
			for _, d := range districts {
				if strings.Contains(strings.ToLower(d.Name), strings.ToLower(district)) {
					suggestion.District = d.Name
					matchedDistrict = true
					break
				}
			}
		}
	}

	// Score: each non-empty input field that resolves contributes 1/3 confidence.
	var scored, total float64
	if province != "" {
		total++
		if matchedProvince != nil {
			scored++
		}
	}
	if city != "" {
		total++
		if matchedCity != nil {
			scored++
		}
	}
	if district != "" {
		total++
		if matchedDistrict {
			scored++
		}
	}

	if total > 0 {
		suggestion.Confidence = scored / total
	}
	return suggestion
}
