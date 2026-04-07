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

// PlaceSuggestion is a single autocomplete prediction.
type PlaceSuggestion struct {
	PlaceID     string `json:"placeId"`
	Description string `json:"description"`
	MainText    string `json:"mainText"`
	SecondText  string `json:"secondText"`
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
