package services

import (
	"context"
	"strings"

	"github.com/fiando/propti/backend/internal/models"
	"github.com/fiando/propti/backend/internal/utils"
)

type SearchIntentParser interface {
	ParseSearchIntent(ctx context.Context, query string) (*models.SearchIntent, error)
}

type SearchIntentLocationResolver interface {
	NormalizeSuggestion(province, city, district string) models.ParsedLocationSuggestion
}

type SearchIntentService struct {
	parser          SearchIntentParser
	locationCatalog SearchIntentLocationResolver
}

func NewSearchIntentService(parser SearchIntentParser, locationCatalog SearchIntentLocationResolver) *SearchIntentService {
	return &SearchIntentService{
		parser:          parser,
		locationCatalog: locationCatalog,
	}
}

func (s *SearchIntentService) ParseIntent(ctx context.Context, query string) (*models.SearchIntentResponse, error) {
	if s.parser == nil {
		return nil, utils.NewAppError(503, "smart search unavailable")
	}

	intent, err := s.parser.ParseSearchIntent(ctx, query)
	if err != nil {
		utils.LogError("parse search intent", err)
		return nil, utils.ErrInternal
	}

	params, metadata := s.Normalize(intent)

	normalized := *intent
	normalized.Query = query
	normalized.ListingType = string(params.ListingType)
	normalized.Province = params.Province
	normalized.City = params.City
	normalized.LegalStatus = params.LegalStatus
	normalized.Amenities = append([]string(nil), params.Amenities...)
	normalized.SortBy = params.SortBy
	normalized.KeywordQuery = params.Query

	return &models.SearchIntentResponse{
		SearchParams: params,
		Normalized:   normalized,
		Metadata:     metadata,
	}, nil
}

func (s *SearchIntentService) Normalize(intent *models.SearchIntent) (models.ListingSearchParams, models.SearchIntentMetadata) {
	if intent == nil {
		return models.ListingSearchParams{}, models.SearchIntentMetadata{}
	}

	params := models.ListingSearchParams{
		Query:           strings.TrimSpace(intent.KeywordQuery),
		PriceMin:        intent.PriceMin,
		PriceMax:        intent.PriceMax,
		Bedrooms:        intent.Bedrooms,
		Bathrooms:       intent.Bathrooms,
		BuildingAreaMin: intent.BuildingAreaMin,
		BuildingAreaMax: intent.BuildingAreaMax,
		LandAreaMin:     intent.LandAreaMin,
		LandAreaMax:     intent.LandAreaMax,
	}

	switch normalizeText(intent.ListingType) {
	case "sell", "jual", "dijual":
		params.ListingType = models.ListingTypeSell
	case "rent", "sewa", "disewa", "kontrak":
		params.ListingType = models.ListingTypeRent
	}

	params.LegalStatus = normalizeLegalStatus(intent.LegalStatus)
	params.Amenities = normalizeAmenityIDs(intent.Amenities)
	params.SortBy = normalizeSortBy(intent.SortBy)

	metadata := models.SearchIntentMetadata{}
	province := normalizeProvinceAlias(intent.Province)
	city := normalizeCityAlias(intent.City)
	if s.locationCatalog != nil && (province != "" || city != "") {
		suggestion := s.locationCatalog.NormalizeSuggestion(province, city, "")
		if suggestion.Province != "" {
			params.Province = suggestion.Province
		}
		if suggestion.City != "" {
			params.City = suggestion.City
		}
		metadata.LocationResolved = suggestion.Province != "" || suggestion.City != ""
	} else {
		params.Province = province
		params.City = city
	}

	return params, metadata
}

var provinceAliases = map[string]string{
	"diy":            "DI Yogyakarta",
	"di yogyakarta":  "DI Yogyakarta",
	"d.i yogyakarta": "DI Yogyakarta",
	"d.i. yogyakarta":"DI Yogyakarta",
	"jogja":          "DI Yogyakarta",
	"yogya":          "DI Yogyakarta",
	"yogyakarta":     "DI Yogyakarta",
}

var cityAliases = map[string]string{
	"gunung kidul": "Gunungkidul",
	"gunungkidul":  "Gunungkidul",
}

func normalizeProvinceAlias(value string) string {
	normalized := normalizeText(value)
	if normalized == "" {
		return ""
	}
	if alias, ok := provinceAliases[normalized]; ok {
		return alias
	}
	return strings.TrimSpace(value)
}

func normalizeCityAlias(value string) string {
	normalized := normalizeText(value)
	if normalized == "" {
		return ""
	}
	if alias, ok := cityAliases[normalized]; ok {
		return alias
	}
	return strings.TrimSpace(value)
}

func normalizeLegalStatus(value string) string {
	switch normalizeText(value) {
	case "shm", "sertifikat hak milik":
		return "SHM"
	case "hgb", "hak guna bangunan":
		return "HGB"
	case "shsrs", "strata title":
		return "SHSRS"
	case "girik", "letter c":
		return "Girik"
	case "ajb", "akta jual beli":
		return "AJB"
	case "lainnya", "lainnya lain":
		return "Lainnya"
	default:
		return ""
	}
}

var amenityAliases = map[string]string{
	"ruang_tamu":            "ruang_tamu",
	"ruang tamu":            "ruang_tamu",
	"ruang_keluarga":        "ruang_keluarga",
	"ruang keluarga":        "ruang_keluarga",
	"dapur":                 "dapur",
	"carport":               "carport",
	"garasi":                "garasi",
	"taman":                 "taman",
	"teras":                 "teras",
	"kanopi":                "kanopi",
	"kolam_renang":          "kolam_renang",
	"kolam renang":          "kolam_renang",
	"balkon":                "balkon",
	"gudang":                "gudang",
	"ruang_makan":           "ruang_makan",
	"ruang makan":           "ruang_makan",
	"ruang_kerja":           "ruang_kerja",
	"ruang kerja":           "ruang_kerja",
	"ruang_cuci":            "ruang_cuci",
	"ruang cuci":            "ruang_cuci",
	"kamar_pembantu":        "kamar_pembantu",
	"kamar pembantu":        "kamar_pembantu",
	"kamar_mandi_pembantu":  "kamar_mandi_pembantu",
	"km pembantu":           "kamar_mandi_pembantu",
	"tempat_jemuran":        "tempat_jemuran",
	"area jemur":            "tempat_jemuran",
	"pantry":                "pantry",
	"ac":                    "ac",
	"water_heater":          "water_heater",
	"water heater":          "water_heater",
	"kitchen_set":           "kitchen_set",
	"kitchen set":           "kitchen_set",
	"furnished":             "furnished",
	"fully furnished":       "furnished",
	"semi_furnished":        "semi_furnished",
	"semi furnished":        "semi_furnished",
	"internet_wifi":         "internet_wifi",
	"internet / wifi":       "internet_wifi",
	"internet":              "internet_wifi",
	"wifi":                  "internet_wifi",
	"tv_kabel":              "tv_kabel",
	"tv kabel":              "tv_kabel",
	"pompa_air":             "pompa_air",
	"pompa air":             "pompa_air",
	"sumur_bor":             "sumur_bor",
	"sumur bor":             "sumur_bor",
	"pdams":                 "pdams",
	"air pdam":              "pdams",
	"listrik_3_phase":       "listrik_3_phase",
	"listrik 3 phase":       "listrik_3_phase",
	"keamanan_24jam":        "keamanan_24jam",
	"keamanan 24 jam":       "keamanan_24jam",
	"cctv":                  "cctv",
	"one_gate_system":       "one_gate_system",
	"one gate system":       "one_gate_system",
	"akses_kartu":           "akses_kartu",
	"access card":           "akses_kartu",
	"lift":                  "lift",
	"lobi":                  "lobi",
	"lobi / reception":      "lobi",
	"gym":                   "gym",
	"gym / fitness":         "gym",
	"clubhouse":             "clubhouse",
	"masjid":                "masjid",
	"masjid / mushola":      "masjid",
	"playground":            "playground",
	"jogging_track":         "jogging_track",
	"jogging track":         "jogging_track",
	"lapangan_olahraga":     "lapangan_olahraga",
	"lapangan olahraga":     "lapangan_olahraga",
	"function_room":         "function_room",
	"function room":         "function_room",
	"loading_dock":          "loading_dock",
	"loading dock":          "loading_dock",
	"akses_container":       "akses_container",
	"akses container":       "akses_container",
	"akses_truk":            "akses_truk",
	"akses truk":            "akses_truk",
	"fire_safety":           "fire_safety",
	"fire safety":           "fire_safety",
	"dekat_tol":             "dekat_tol",
	"dekat tol":             "dekat_tol",
	"jalan_lebar":           "jalan_lebar",
	"akses jalan lebar":     "jalan_lebar",
}

func normalizeAmenityIDs(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := normalizeText(value)
		if normalized == "" {
			continue
		}
		if amenityID, ok := amenityAliases[normalized]; ok {
			unique[amenityID] = struct{}{}
		}
	}

	normalizedValues := make([]string, 0, len(unique))
	for _, value := range values {
		amenityID, ok := amenityAliases[normalizeText(value)]
		if !ok {
			continue
		}
		if _, exists := unique[amenityID]; !exists {
			continue
		}
		normalizedValues = append(normalizedValues, amenityID)
		delete(unique, amenityID)
	}

	if len(normalizedValues) == 0 {
		return nil
	}

	return normalizedValues
}

func normalizeSortBy(value string) string {
	switch normalizeText(value) {
	case "price_asc", "harga terendah", "termurah", "harga termurah":
		return "price_asc"
	case "price_desc", "harga tertinggi", "termahal", "harga termahal":
		return "price_desc"
	case "popular", "populer", "paling banyak dilihat":
		return "popular"
	case "newest", "terbaru":
		return "newest"
	default:
		return ""
	}
}

func normalizeText(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.NewReplacer("-", " ", "_", " ", "/", " ").Replace(normalized)
	return strings.Join(strings.Fields(normalized), " ")
}
