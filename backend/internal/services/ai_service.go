package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/fiando/propti/backend/internal/models"
)

const parseSystemPrompt = `You are an expert real estate assistant for the Indonesian property market.
Your task is to parse free-form Indonesian real estate listing text and extract structured information with clean, buyer-friendly wording.

Extract the following fields from the text and return ONLY valid JSON:
{
  "title": "string - concise listing title",
  "description": "string - cleaned up description",
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
  - Prefer a structure like:
    Deskripsi singkat / opening line
    \n\n✨ Highlight utama
    \n• poin 1
    \n• poin 2
    \n• poin 3
    \n\n📍 Lokasi
    \n<lokasi ringkas>
  - Do not use markdown headings, code fences, or excessive emojis.
  - Keep wording professional and property-focused.
- Format "address" as a clean, readable full address string for a form field:
  - combine street/area/location clues into one natural line
  - remove noisy filler words like "lok", "dkt", "hub", repeated punctuation, or phone numbers
  - preserve meaningful street, district, city, and province details when present
- For "locationSuggestion", infer province/city/district from context clues; leave fields empty rather than guessing
- If a field cannot be determined, use null for numbers, empty string for strings, empty array for arrays
- Set requiresManualReview to true if price, address, or property type cannot be determined
- Return ONLY the JSON object, no markdown, no explanation`

const moderationSystemPrompt = `You are a content moderation assistant for an Indonesian real estate platform.
Analyze the listing content for the following violations:
1. Spam or duplicate/nonsensical content
2. Illegal property (e.g. on protected land, without permits)
3. Fraudulent claims or misleading pricing
4. Inappropriate content unrelated to real estate
5. Price that is unrealistically low or high for the described property

Respond ONLY with valid JSON:
{
  "approved": boolean,
  "reason": "string - explanation if rejected, empty string if approved",
  "flags": ["array of specific issues found"]
}`

const parserModel = "gpt-4o-mini"
const moderationModel = "gpt-4o-mini"

// AIService calls the OpenAI API for listing text parsing and content moderation.
type AIService struct {
	client *openai.Client
}

// NewAIService creates an AIService using the provided OpenAI API key.
func NewAIService(apiKey string) *AIService {
	return &AIService{client: openai.NewClient(apiKey)}
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

// moderationResponse is the raw JSON structure returned by the moderation prompt.
type moderationResponse struct {
	Approved bool     `json:"approved"`
	Reason   string   `json:"reason"`
	Flags    []string `json:"flags"`
}

// ModerateContent checks a listing title + description for policy violations.
func (s *AIService) ModerateContent(ctx context.Context, title, description string) (approved bool, reason string, flags []string, err error) {
	input := fmt.Sprintf("Title: %s\n\nDescription: %s", title, description)

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: moderationModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: moderationSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: input},
		},
		Temperature:    0.0,
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
	})
	if err != nil {
		return false, "", nil, fmt.Errorf("openai moderation request: %w", err)
	}

	if len(resp.Choices) == 0 {
		return false, "", nil, fmt.Errorf("openai returned no choices")
	}

	content := strings.TrimSpace(resp.Choices[0].Message.Content)
	var modResp moderationResponse
	if err := json.Unmarshal([]byte(content), &modResp); err != nil {
		return false, "", nil, fmt.Errorf("unmarshal moderation response: %w", err)
	}
	return modResp.Approved, modResp.Reason, modResp.Flags, nil
}
