package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/fiando/propti/backend/internal/models"
)

const parseSystemPrompt = `You are an expert real estate assistant for the Indonesian property market.
Your task is to parse free-form Indonesian real estate listing text and extract structured information.
Preserve the seller's voice, promotional language, and all selling points — only reformat for readability.

Extract the following fields from the text and return ONLY valid JSON:
{
  "title": "string - concise listing title",
  "description": "string - reformatted description that preserves all original selling points, tone, and promotional language",
  "price": number - price in IDR (convert M/jt/miliar/milyard to full number),
  "priceUnit": "string - 'total' for buy, 'monthly' or 'yearly' for rent",
  "propertyDetails": {
    "landArea": number - in m²,
    "buildingArea": number - in m²,
    "bedrooms": integer,
    "bathrooms": integer,
    "frontWidth": number - in meters,
    "orientation": "string - e.g. utara/selatan/timur/barat",
    "legalStatus": "string - SHM/HGB/AJB/SHGB",
    "powerConsumption": "string - e.g. 1300W/2200W",
    "amenities": ["array of amenities"]
  },
  "address": "string - cleaned full address text for the form field, preserving all useful location detail without inventing unsupported facts",
  "locationSuggestion": {
    "province": "string - suggested province name, leave empty if uncertain",
    "city": "string - suggested city/kabupaten name, leave empty if uncertain",
    "district": "string - suggested kecamatan/district name, leave empty if uncertain",
    "normalizedAddress": "string - best-effort normalized full address, leave empty if uncertain",
    "confidence": number between 0.0 and 1.0 indicating location suggestion confidence
  },
  "confidence": number between 0.0 and 1.0 indicating overall parse confidence,
  "requiresManualReview": boolean - true if critical info is missing or ambiguous,
  "warnings": ["array of warning strings for missing or ambiguous fields"]
}

Rules:
- Convert price shorthand: 1M/1 juta = 1000000, 1 miliar/1M (if context says miliar) = 1000000000
- If bedrooms/bathrooms use KT/KM notation: KT = kamar tidur (bedrooms), KM = kamar mandi (bathrooms)
- LT = luas tanah (land area), LB = luas bangunan (building area)
- Format "title" as clean headline-style Indonesian copy:
  - readable, concise, and natural
  - no ALL CAPS
  - no spammy wording
  - include key value points like property type, bedroom count, area, and main location when available
- Format "description" as clean multiline Indonesian copy using short sections and emoji bullets when helpful.
  - Use newline characters in the JSON string to separate lines.
  - Keep it easy to scan on mobile.
  - PRESERVE the original promotional tone, urgency phrases, and marketing language (e.g. "Pas banget nih...", "Monggo sebelum keduluan", "Harga sangat murah").
  - KEEP all selling points: bonuses/furnishing mentions, price comparisons, proximity highlights, urgency cues — these are valuable to buyers.
  - Only reformat for readability (fix typos, structure with bullets, remove duplicate info), never remove meaningful content.
  - Prefer a structure like:
    <opening / hook from original text if present>
    \n\n✨ Highlight utama
    \n• poin 1
    \n• poin 2
    \n• poin 3
    \n\n📍 Lokasi
    \n<lokasi ringkas>
  - Do not use markdown headings, code fences, or excessive emojis.
  - Do not start with a generic "Deskripsi singkat:" prefix.
- Format "address" as a clean, readable full address string for a form field:
  - combine street/area/location clues into one natural line
  - remove noisy filler words like "lok", "dkt", "hub", repeated punctuation, or phone numbers
  - preserve meaningful street, district, city, and province details when present
- For "locationSuggestion", infer province/city/district from context clues; leave fields empty rather than guessing
- If a field cannot be determined, use null for numbers, empty string for strings, empty array for arrays
- Set requiresManualReview to true if price, address, or property type cannot be determined
- Return ONLY the JSON object, no markdown, no explanation`

const parserModel = "gpt-4o-mini"
const searchIntentModel = "gpt-4.1-nano"
const openAIModerationModel = "omni-moderation-latest"
const openAIModerationBaseURL = "https://api.openai.com"
const searchIntentMaxTokens = 250

const searchIntentSystemPrompt = `You extract Indonesian buyer property-search intent into JSON for an existing filter sidebar.
Return ONLY valid JSON with this shape:
{
  "query": "string - copy of the buyer sentence",
  "keywordQuery": "string - leftover keyword intent not already captured by structured filters, or empty string",
  "listingType": "sell | rent | empty string",
  "province": "string",
  "city": "string",
  "priceMin": number,
  "priceMax": number,
  "bedrooms": integer,
  "bathrooms": integer,
  "buildingAreaMin": number,
  "buildingAreaMax": number,
  "landAreaMin": number,
  "landAreaMax": number,
  "legalStatus": "SHM | HGB | SHSRS | Girik | AJB | Lainnya | empty string",
  "amenities": ["array of amenity ids from the supported list"],
  "sortBy": "newest | price_asc | price_desc | popular | empty string",
  "confidence": number
}

Supported amenity ids:
ruang_tamu, ruang_keluarga, dapur, carport, garasi, taman, teras, kanopi, kolam_renang, balkon, gudang, ruang_makan, ruang_kerja, ruang_cuci, kamar_pembantu, kamar_mandi_pembantu, tempat_jemuran, pantry, ac, water_heater, kitchen_set, furnished, semi_furnished, internet_wifi, tv_kabel, pompa_air, sumur_bor, pdams, listrik_3_phase, keamanan_24jam, cctv, one_gate_system, akses_kartu, lift, lobi, gym, clubhouse, masjid, playground, jogging_track, lapangan_olahraga, function_room, loading_dock, akses_container, akses_truk, fire_safety, dekat_tol, jalan_lebar

Rules:
- Use only the exact enum values above for listingType, legalStatus, amenities, and sortBy.
- Map "harga termurah" or similar to "price_asc", "harga tertinggi" to "price_desc", "terbaru" to "newest", "paling populer" or "paling banyak dilihat" to "popular".
- Convert rupiah shorthand into full numbers.
- KT = bedrooms, KM = bathrooms, LB = building area, LT = land area.
- Leave string fields empty and numeric fields 0 when not specified.
- Prefer structured extraction over prose. No explanation, no markdown.`

// AIService calls the OpenAI API for listing text parsing and content moderation.
type AIService struct {
	client            *openai.Client
	httpClient        *http.Client
	moderationAPIKey  string
	moderationBaseURL string
}

// NewAIService creates an AIService using the provided OpenAI API key.
func NewAIService(apiKey string) *AIService {
	return &AIService{
		client:            openai.NewClient(apiKey),
		httpClient:        http.DefaultClient,
		moderationAPIKey:  apiKey,
		moderationBaseURL: openAIModerationBaseURL,
	}
}

// ParseListingText sends raw Indonesian listing text to the parser model and returns structured data.
func (s *AIService) ParseListingText(ctx context.Context, text string) (*models.ParsedListing, error) {
	resp, err := s.client.CreateChatCompletion(ctx, buildParseChatCompletionRequest(text))
	if err != nil {
		return nil, fmt.Errorf("openai parse request: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	var parsed models.ParsedListing
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal parsed listing: %w", err)
	}
	return &parsed, nil
}

func (s *AIService) ParseSearchIntent(ctx context.Context, query string) (*models.SearchIntent, error) {
	resp, err := s.client.CreateChatCompletion(ctx, buildSearchIntentChatCompletionRequest(query))
	if err != nil {
		return nil, fmt.Errorf("openai search intent request: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	var parsed models.SearchIntent
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, fmt.Errorf("unmarshal search intent: %w", err)
	}
	return &parsed, nil
}

func buildParseChatCompletionRequest(text string) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: parserModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: parseSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: text},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
	}
}

func buildSearchIntentChatCompletionRequest(query string) openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: searchIntentModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: searchIntentSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: query},
		},
		Temperature:    0,
		MaxTokens:      searchIntentMaxTokens,
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
	}
}

// ModerateContent checks a listing title + description for policy violations.
func (s *AIService) ModerateContent(ctx context.Context, title, description string) (approved bool, reason string, flags []string, err error) {
	input := fmt.Sprintf("Title: %s\n\nDescription: %s", title, description)
	return s.moderateHarmfulText(ctx, input)
}

func (s *AIService) ModerateImages(ctx context.Context, images [][]byte) (approved bool, reason string, flags []string, err error) {
	for idx, imageBytes := range images {
		input := []openAIModerationInputItem{
			{
				Type: "image_url",
				ImageURL: &openAIModerationImageURL{
					URL: buildModerationImageDataURL(imageBytes),
				},
			},
		}
		result, err := s.createModeration(ctx, input)
		if err != nil {
			return false, "", nil, fmt.Errorf("openai image moderation request %d: %w", idx+1, err)
		}
		if !result.Flagged {
			continue
		}

		flags = moderationCategoryFlags(result.Categories)
		return false, fmt.Sprintf("Gambar %d terdeteksi mengandung konten yang tidak pantas untuk iklan properti", idx+1), flags, nil
	}

	return true, "", nil, nil
}

func (s *AIService) moderateHarmfulText(ctx context.Context, input string) (approved bool, reason string, flags []string, err error) {
	result, err := s.createModeration(ctx, input)
	if err != nil {
		return false, "", nil, fmt.Errorf("openai text moderation request: %w", err)
	}
	if !result.Flagged {
		return true, "", nil, nil
	}

	flags = moderationCategoryFlags(result.Categories)
	return false, "Konten teks terdeteksi mengandung materi yang tidak aman untuk platform properti", flags, nil
}

type openAIModerationRequest struct {
	Model string `json:"model"`
	Input any    `json:"input"`
}

type openAIModerationInputItem struct {
	Type     string                    `json:"type"`
	Text     string                    `json:"text,omitempty"`
	ImageURL *openAIModerationImageURL `json:"image_url,omitempty"`
}

type openAIModerationImageURL struct {
	URL string `json:"url"`
}

type openAIModerationResponse struct {
	Results []openAIModerationResult `json:"results"`
}

type openAIModerationResult struct {
	Flagged    bool            `json:"flagged"`
	Categories map[string]bool `json:"categories"`
}

func (s *AIService) createModeration(ctx context.Context, input any) (*openAIModerationResult, error) {
	if strings.TrimSpace(s.moderationAPIKey) == "" {
		return nil, fmt.Errorf("openai moderation api key not configured")
	}

	httpClient := s.httpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	body, err := json.Marshal(openAIModerationRequest{
		Model: openAIModerationModel,
		Input: input,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal moderation request: %w", err)
	}

	url := strings.TrimRight(s.moderationBaseURL, "/") + "/v1/moderations"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build moderation request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.moderationAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send moderation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("moderation request failed with status %d", resp.StatusCode)
	}

	var moderationResp openAIModerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&moderationResp); err != nil {
		return nil, fmt.Errorf("decode moderation response: %w", err)
	}
	if len(moderationResp.Results) == 0 {
		return nil, fmt.Errorf("openai moderation returned no results")
	}

	return &moderationResp.Results[0], nil
}

func moderationCategoryFlags(categories map[string]bool) []string {
	flags := make([]string, 0, len(categories))
	for category, flagged := range categories {
		if flagged {
			flags = append(flags, category)
		}
	}
	sort.Strings(flags)
	return flags
}

func buildModerationImageDataURL(imageBytes []byte) string {
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(imageBytes)
}
