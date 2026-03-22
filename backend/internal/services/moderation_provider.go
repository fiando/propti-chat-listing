package services

import (
	"context"
	"fmt"
	"strings"
)

type imageModerationProvider string

const (
	imageModerationProviderOpenAI      imageModerationProvider = "openai"
	imageModerationProviderRekognition imageModerationProvider = "rekognition"
)

func parseImageModerationProvider(raw string) (imageModerationProvider, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(imageModerationProviderOpenAI):
		return imageModerationProviderOpenAI, nil
	case string(imageModerationProviderRekognition):
		return imageModerationProviderRekognition, nil
	default:
		return "", fmt.Errorf("unsupported image moderation provider %q", raw)
	}
}

func NewImageModerator(ctx context.Context, providerValue, openAIAPIKey string) (ImageModerator, error) {
	provider, err := parseImageModerationProvider(providerValue)
	if err != nil {
		return nil, err
	}

	switch provider {
	case imageModerationProviderOpenAI:
		if strings.TrimSpace(openAIAPIKey) == "" {
			return nil, fmt.Errorf("openai api key is required when IMAGE_MODERATION_PROVIDER=openai")
		}
		return NewAIService(openAIAPIKey), nil
	case imageModerationProviderRekognition:
		return NewRekognitionImageModerator(ctx)
	default:
		return nil, fmt.Errorf("unsupported image moderation provider %q", provider)
	}
}
