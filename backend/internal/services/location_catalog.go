package services

import (
	"encoding/json"
	"io"
	"strings"
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
	var results []Province
	for _, p := range c.provinces {
		if strings.Contains(strings.ToLower(p.Name), q) {
			results = append(results, p)
		}
	}
	return results
}

func (c *LocationCatalog) SearchCities(provinceID, query string) []City {
	q := strings.ToLower(query)
	var results []City
	for _, city := range c.citiesByProvince[provinceID] {
		if strings.Contains(strings.ToLower(city.Name), q) {
			results = append(results, city)
		}
	}
	return results
}

func (c *LocationCatalog) SearchDistricts(cityID, query string) []District {
	q := strings.ToLower(query)
	var results []District
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
