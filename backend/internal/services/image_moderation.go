package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rekognitiontypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

type rekognitionAPI interface {
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
