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
Your task is to parse free-form Indonesian real estate listing text and extract structured information.

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
  "address": "string - raw address text exactly as found in the listing, do not normalize",
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
- Keep "address" as the raw extracted text from the listing without any normalization
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

// AIService calls the OpenAI API for listing text parsing and content moderation.
type AIService struct {
	client *openai.Client
}

// NewAIService creates an AIService using the provided OpenAI API key.
func NewAIService(apiKey string) *AIService {
	return &AIService{client: openai.NewClient(apiKey)}
}

// ParseListingText sends raw Indonesian listing text to gpt-5-mini and returns structured data.
func (s *AIService) ParseListingText(ctx context.Context, text string) (*models.ParsedListing, error) {
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: "gpt-5-mini",
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: parseSystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: text},
		},
		Temperature:    0.1,
		ResponseFormat: &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject},
	})
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
		Model: openai.GPT4oMini,
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
