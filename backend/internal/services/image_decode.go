package services

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func decodeListingImage(raw string) ([]byte, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("empty image payload")
	}

	if comma := strings.Index(trimmed, ","); strings.HasPrefix(trimmed, "data:") && comma >= 0 {
		trimmed = trimmed[comma+1:]
	}

	data, err := base64.StdEncoding.DecodeString(trimmed)
	if err == nil {
		return data, nil
	}

	data, err = base64.RawStdEncoding.DecodeString(trimmed)
	if err == nil {
		return data, nil
	}

	return nil, fmt.Errorf("invalid base64 image data")
}
