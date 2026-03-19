package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rekognitiontypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

var propertyRelevantLabels = map[string]struct{}{
	"apartment":       {},
	"architecture":    {},
	"balcony":         {},
	"bathroom":        {},
	"bedroom":         {},
	"building":        {},
	"cottage":         {},
	"dining room":     {},
	"door":            {},
	"facade":          {},
	"garage":          {},
	"house":           {},
	"housing":         {},
	"interior design": {},
	"kitchen":         {},
	"living room":     {},
	"patio":           {},
	"pool":            {},
	"porch":           {},
	"property":        {},
	"real estate":     {},
	"resort":          {},
	"staircase":       {},
	"swimming pool":   {},
	"villa":           {},
	"window":          {},
	"yard":            {},
}

type rekognitionAPI interface {
	DetectLabels(ctx context.Context, params *rekognition.DetectLabelsInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectLabelsOutput, error)
	DetectModerationLabels(ctx context.Context, params *rekognition.DetectModerationLabelsInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectModerationLabelsOutput, error)
}

type RekognitionImageModerator struct {
	client rekognitionAPI
}

func NewRekognitionImageModerator(ctx context.Context) (*RekognitionImageModerator, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config for rekognition: %w", err)
	}

	return &RekognitionImageModerator{
		client: rekognition.NewFromConfig(cfg),
	}, nil
}

func (m *RekognitionImageModerator) ModerateImages(ctx context.Context, images [][]byte) (bool, string, []string, error) {
	for idx, bytes := range images {
		moderationResult, err := m.client.DetectModerationLabels(ctx, &rekognition.DetectModerationLabelsInput{
			Image: &rekognitiontypes.Image{Bytes: bytes},
		})
		if err != nil {
			return false, "", nil, fmt.Errorf("detect moderation labels for image %d: %w", idx+1, err)
		}
		if len(moderationResult.ModerationLabels) > 0 {
			flags := make([]string, 0, len(moderationResult.ModerationLabels))
			for _, label := range moderationResult.ModerationLabels {
				if label.Name != nil {
					flags = append(flags, *label.Name)
				}
			}
			return false, fmt.Sprintf("Gambar %d terdeteksi mengandung konten yang tidak pantas untuk iklan properti", idx+1), flags, nil
		}

		labelsResult, err := m.client.DetectLabels(ctx, &rekognition.DetectLabelsInput{
			Image:     &rekognitiontypes.Image{Bytes: bytes},
			MaxLabels: aws.Int32(15),
		})
		if err != nil {
			return false, "", nil, fmt.Errorf("detect labels for image %d: %w", idx+1, err)
		}

		if !looksLikeProperty(labelsResult.Labels) {
			return false, fmt.Sprintf("Gambar %d tidak terlihat relevan dengan properti yang dijual atau disewakan", idx+1), collectLabelNames(labelsResult.Labels), nil
		}
	}

	return true, "", nil, nil
}

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

const propertyLabelMinConfidence = float32(75.0)

// looksLikeProperty checks only the DIRECT label name (not parent/alias/category
// expansions) and requires at least 75% confidence to avoid false positives from
// unrelated images (e.g. game screenshots whose parent category is "Indoor").
func looksLikeProperty(labels []rekognitiontypes.Label) bool {
	for _, label := range labels {
		if label.Name == nil || label.Confidence == nil {
			continue
		}
		if *label.Confidence < propertyLabelMinConfidence {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(*label.Name))
		if _, ok := propertyRelevantLabels[name]; ok {
			return true
		}
	}
	return false
}

func collectLabelNames(labels []rekognitiontypes.Label) []string {
	seen := make(map[string]struct{})
	names := make([]string, 0, len(labels))

	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return
		}
		if _, exists := seen[value]; exists {
			return
		}
		seen[value] = struct{}{}
		names = append(names, value)
	}

	for _, label := range labels {
		if label.Name != nil {
			add(*label.Name)
		}
		for _, parent := range label.Parents {
			if parent.Name != nil {
				add(*parent.Name)
			}
		}
		for _, alias := range label.Aliases {
			if alias.Name != nil {
				add(*alias.Name)
			}
		}
		for _, category := range label.Categories {
			if category.Name != nil {
				add(*category.Name)
			}
		}
	}

	return names
}
