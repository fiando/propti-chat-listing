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
func (c *LocationCatalog) NormalizeSuggestion(province, city, district string) models.ParsedLocationSuggestion {
	suggestion := models.ParsedLocationSuggestion{
		Province: province,
		City:     city,
		District: district,
	}

	var matchedProvince *Province
	if province != "" {
		for i, p := range c.provinces {
			if strings.EqualFold(p.Name, province) ||
				strings.Contains(strings.ToLower(p.Name), strings.ToLower(province)) {
				matchedProvince = &c.provinces[i]
				suggestion.Province = p.Name
				break
			}
		}
	}

	var matchedCity *City
	if city != "" && matchedProvince != nil {
		for i, ct := range c.citiesByProvince[matchedProvince.ID] {
			if strings.EqualFold(ct.Name, city) ||
				strings.Contains(strings.ToLower(ct.Name), strings.ToLower(city)) {
				matchedCity = &c.citiesByProvince[matchedProvince.ID][i]
				suggestion.City = ct.Name
				break
			}
		}
	}

	var matchedDistrict bool
	if district != "" && matchedCity != nil {
		for _, d := range c.districtsByCity[matchedCity.ID] {
			if strings.EqualFold(d.Name, district) ||
				strings.Contains(strings.ToLower(d.Name), strings.ToLower(district)) {
				suggestion.District = d.Name
				matchedDistrict = true
				break
			}
		}
	}

	// Score: each non-empty field that resolves contributes 1/3 confidence.
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
