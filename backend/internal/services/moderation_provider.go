package services

import (
	"fmt"
	"strings"
)

func NewImageModerator(openAIAPIKey string) (ImageModerator, error) {
	if strings.TrimSpace(openAIAPIKey) == "" {
		return nil, fmt.Errorf("openai api key is required")
	}
	return NewAIService(openAIAPIKey), nil
}
