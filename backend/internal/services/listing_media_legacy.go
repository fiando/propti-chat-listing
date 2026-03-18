package services

import (
	"encoding/base64"
	"strings"

	"github.com/fiando/propti/backend/internal/models"
)

func containsLegacyImagePayload(images []string) bool {
	for _, image := range images {
		if isLegacyImagePayload(image) {
			return true
		}
	}
	return false
}

func isLegacyImagePayload(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}
	if strings.HasPrefix(trimmed, "data:image/") {
		return true
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return false
	}
	if len(trimmed) < 32 {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(trimmed)
	return err == nil
}

func legacyImageValues(images models.ImageEntries) []string {
	return images.LegacyValues()
}
