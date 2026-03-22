package models

type SearchIntentRequest struct {
	Query string `json:"query"`
}

type SearchIntent struct {
	Query           string      `json:"query"`
	KeywordQuery    string      `json:"keywordQuery"`
	ListingType     string      `json:"listingType"`
	Province        string      `json:"province"`
	City            string      `json:"city"`
	PriceMin        float64     `json:"priceMin"`
	PriceMax        float64     `json:"priceMax"`
	Bedrooms        int         `json:"bedrooms"`
	Bathrooms       int         `json:"bathrooms"`
	BuildingAreaMin float64     `json:"buildingAreaMin"`
	BuildingAreaMax float64     `json:"buildingAreaMax"`
	LandAreaMin     float64     `json:"landAreaMin"`
	LandAreaMax     float64     `json:"landAreaMax"`
	LegalStatus     string      `json:"legalStatus"`
	Amenities       []string    `json:"amenities"`
	SortBy          string      `json:"sortBy"`
	Confidence      float64     `json:"confidence"`
}

type SearchIntentMetadata struct {
	LocationResolved bool `json:"locationResolved"`
}

type SearchIntentResponse struct {
	SearchParams ListingSearchParams  `json:"searchParams"`
	Normalized   SearchIntent         `json:"normalized"`
	Metadata     SearchIntentMetadata `json:"metadata"`
}
