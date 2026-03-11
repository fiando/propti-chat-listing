package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// GoogleMapsService calls the Google Maps Platform APIs.
type GoogleMapsService struct {
	apiKey string
	client *http.Client
}

// NewGoogleMapsService creates a GoogleMapsService with the given API key.
func NewGoogleMapsService(apiKey string) *GoogleMapsService {
	return &GoogleMapsService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// NewGoogleMapsServiceFromEnv creates a GoogleMapsService using GOOGLE_MAPS_API_KEY env var.
func NewGoogleMapsServiceFromEnv() *GoogleMapsService {
	return NewGoogleMapsService(os.Getenv("GOOGLE_MAPS_API_KEY"))
}

// GeocodedLocation holds the result of a geocode or reverse-geocode request.
type GeocodedLocation struct {
	FormattedAddress string  `json:"formattedAddress"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	PlaceID          string  `json:"placeId"`
	Province         string  `json:"province"`
	City             string  `json:"city"`
	District         string  `json:"district"`
}

// PlaceSuggestion is a single autocomplete prediction.
type PlaceSuggestion struct {
	PlaceID     string `json:"placeId"`
	Description string `json:"description"`
	MainText    string `json:"mainText"`
	SecondText  string `json:"secondText"`
}

// PlaceDetails holds the full details of a Google Place.
type PlaceDetails struct {
	PlaceID          string  `json:"placeId"`
	FormattedAddress string  `json:"formattedAddress"`
	Latitude         float64 `json:"latitude"`
	Longitude        float64 `json:"longitude"`
	Province         string  `json:"province"`
	City             string  `json:"city"`
	District         string  `json:"district"`
}

// geocodeResponse mirrors the Google Maps Geocoding API response.
type geocodeResponse struct {
	Results []struct {
		FormattedAddress string `json:"formatted_address"`
		PlaceID          string `json:"place_id"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
	} `json:"results"`
	Status string `json:"status"`
}

// autocompleteResponse mirrors the Places Autocomplete API response.
type autocompleteResponse struct {
	Predictions []struct {
		PlaceID          string `json:"place_id"`
		Description      string `json:"description"`
		StructuredFormat struct {
			MainText      struct{ Text string `json:"text"` } `json:"main_text"`
			SecondaryText struct{ Text string `json:"text"` } `json:"secondary_text"`
		} `json:"structured_formatting"`
	} `json:"predictions"`
	Status string `json:"status"`
}

// placeDetailsResponse mirrors the Places Details API response.
type placeDetailsResponse struct {
	Result struct {
		PlaceID          string `json:"place_id"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		AddressComponents []struct {
			LongName string   `json:"long_name"`
			Types    []string `json:"types"`
		} `json:"address_components"`
	} `json:"result"`
	Status string `json:"status"`
}

// GeocodeAddress converts a free-form address string into coordinates.
func (s *GoogleMapsService) GeocodeAddress(ctx context.Context, address string) (*GeocodedLocation, error) {
	params := url.Values{}
	params.Set("address", address)
	params.Set("key", s.apiKey)
	params.Set("region", "id") // bias towards Indonesia

	apiURL := "https://maps.googleapis.com/maps/api/geocode/json?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("geocode request: %w", err)
	}
	defer resp.Body.Close()

	var data geocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode geocode response: %w", err)
	}
	if data.Status != "OK" || len(data.Results) == 0 {
		return nil, fmt.Errorf("geocode status: %s", data.Status)
	}

	r := data.Results[0]
	loc := &GeocodedLocation{
		FormattedAddress: r.FormattedAddress,
		Latitude:         r.Geometry.Location.Lat,
		Longitude:        r.Geometry.Location.Lng,
		PlaceID:          r.PlaceID,
	}
	extractAdminComponents(r.AddressComponents, loc)
	return loc, nil
}

// ReverseGeocode converts lat/lng into a GeocodedLocation.
func (s *GoogleMapsService) ReverseGeocode(ctx context.Context, lat, lng float64) (*GeocodedLocation, error) {
	params := url.Values{}
	params.Set("latlng", fmt.Sprintf("%f,%f", lat, lng))
	params.Set("key", s.apiKey)

	apiURL := "https://maps.googleapis.com/maps/api/geocode/json?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("reverse geocode request: %w", err)
	}
	defer resp.Body.Close()

	var data geocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode reverse geocode response: %w", err)
	}
	if data.Status != "OK" || len(data.Results) == 0 {
		return nil, fmt.Errorf("reverse geocode status: %s", data.Status)
	}

	r := data.Results[0]
	loc := &GeocodedLocation{
		FormattedAddress: r.FormattedAddress,
		Latitude:         lat,
		Longitude:        lng,
		PlaceID:          r.PlaceID,
	}
	extractAdminComponents(r.AddressComponents, loc)
	return loc, nil
}

// GetPlaceSuggestions calls the Places Autocomplete API.
func (s *GoogleMapsService) GetPlaceSuggestions(ctx context.Context, input string) ([]PlaceSuggestion, error) {
	params := url.Values{}
	params.Set("input", input)
	params.Set("key", s.apiKey)
	params.Set("components", "country:id")
	params.Set("types", "geocode")

	apiURL := "https://maps.googleapis.com/maps/api/place/autocomplete/json?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("autocomplete request: %w", err)
	}
	defer resp.Body.Close()

	var data autocompleteResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode autocomplete response: %w", err)
	}

	suggestions := make([]PlaceSuggestion, 0, len(data.Predictions))
	for _, p := range data.Predictions {
		suggestions = append(suggestions, PlaceSuggestion{
			PlaceID:     p.PlaceID,
			Description: p.Description,
			MainText:    p.StructuredFormat.MainText.Text,
			SecondText:  p.StructuredFormat.SecondaryText.Text,
		})
	}
	return suggestions, nil
}

// GetPlaceDetails retrieves full details for a given place ID.
func (s *GoogleMapsService) GetPlaceDetails(ctx context.Context, placeID string) (*PlaceDetails, error) {
	params := url.Values{}
	params.Set("place_id", placeID)
	params.Set("key", s.apiKey)
	params.Set("fields", "place_id,formatted_address,geometry,address_components")

	apiURL := "https://maps.googleapis.com/maps/api/place/details/json?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("place details request: %w", err)
	}
	defer resp.Body.Close()

	var data placeDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode place details response: %w", err)
	}
	if data.Status != "OK" {
		return nil, fmt.Errorf("place details status: %s", data.Status)
	}

	details := &PlaceDetails{
		PlaceID:          data.Result.PlaceID,
		FormattedAddress: data.Result.FormattedAddress,
		Latitude:         data.Result.Geometry.Location.Lat,
		Longitude:        data.Result.Geometry.Location.Lng,
	}

	for _, c := range data.Result.AddressComponents {
		for _, t := range c.Types {
			switch t {
			case "administrative_area_level_1":
				details.Province = c.LongName
			case "administrative_area_level_2":
				details.City = c.LongName
			case "administrative_area_level_3":
				details.District = c.LongName
			}
		}
	}

	return details, nil
}

// extractAdminComponents fills Province/City/District from address_components.
func extractAdminComponents(
	components []struct {
		LongName string   `json:"long_name"`
		Types    []string `json:"types"`
	},
	loc *GeocodedLocation,
) {
	for _, c := range components {
		for _, t := range c.Types {
			switch t {
			case "administrative_area_level_1":
				loc.Province = c.LongName
			case "administrative_area_level_2":
				loc.City = c.LongName
			case "administrative_area_level_3":
				loc.District = c.LongName
			}
		}
	}
}
