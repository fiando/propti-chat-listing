package services

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	rekognitiontypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
)

type fakeRekognitionClient struct {
	moderationOutputs []*rekognition.DetectModerationLabelsOutput
	labelOutputs      []*rekognition.DetectLabelsOutput
	moderationCalls   int
	labelCalls        int
}

func (f *fakeRekognitionClient) DetectLabels(ctx context.Context, params *rekognition.DetectLabelsInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectLabelsOutput, error) {
	if f.labelCalls >= len(f.labelOutputs) {
		f.labelCalls++
		return &rekognition.DetectLabelsOutput{}, nil
	}

	output := f.labelOutputs[f.labelCalls]
	f.labelCalls++
	return output, nil
}

func (f *fakeRekognitionClient) DetectModerationLabels(ctx context.Context, params *rekognition.DetectModerationLabelsInput, optFns ...func(*rekognition.Options)) (*rekognition.DetectModerationLabelsOutput, error) {
	if f.moderationCalls >= len(f.moderationOutputs) {
		f.moderationCalls++
		return &rekognition.DetectModerationLabelsOutput{}, nil
	}

	output := f.moderationOutputs[f.moderationCalls]
	f.moderationCalls++
	return output, nil
}

func TestRekognitionImageModeratorAllowsSupportingPropertyImages(t *testing.T) {
	client := &fakeRekognitionClient{
		moderationOutputs: []*rekognition.DetectModerationLabelsOutput{
			{},
		},
		labelOutputs: []*rekognition.DetectLabelsOutput{
			{
				Labels: buildLabels(95.0, "Furniture", "Chair", "Indoors"),
			},
		},
	}
	moderator := &RekognitionImageModerator{client: client}

	approved, reason, flags, err := moderator.ModerateImages(context.Background(), [][]byte{[]byte("image-bytes")})
	if err != nil {
		t.Fatalf("ModerateImages returned error: %v", err)
	}
	if !approved {
		t.Fatalf("expected supporting property image to be allowed, got rejected with reason %q and flags %#v", reason, flags)
	}
	if reason != "" {
		t.Fatalf("expected no rejection reason, got %q", reason)
	}
	if len(flags) != 0 {
		t.Fatalf("expected no moderation flags, got %#v", flags)
	}
}

func TestRekognitionImageModeratorRejectsUnsafeContent(t *testing.T) {
	label := "Explicit Nudity"
	client := &fakeRekognitionClient{
		moderationOutputs: []*rekognition.DetectModerationLabelsOutput{
			{
				ModerationLabels: []rekognitiontypes.ModerationLabel{
					{Name: &label},
				},
			},
		},
	}
	moderator := &RekognitionImageModerator{client: client}

	approved, reason, flags, err := moderator.ModerateImages(context.Background(), [][]byte{[]byte("image-bytes")})
	if err != nil {
		t.Fatalf("ModerateImages returned error: %v", err)
	}
	if approved {
		t.Fatal("expected unsafe image to be rejected")
	}
	if reason == "" {
		t.Fatal("expected rejection reason for unsafe image")
	}
	if len(flags) != 1 || flags[0] != label {
		t.Fatalf("expected moderation flags to contain %q, got %#v", label, flags)
	}
	if client.labelCalls != 0 {
		t.Fatalf("expected detect labels not to run after unsafe moderation label, got %d calls", client.labelCalls)
	}
}

func buildLabels(confidence float32, names ...string) []rekognitiontypes.Label {
	labels := make([]rekognitiontypes.Label, 0, len(names))
	for _, name := range names {
		value := name
		conf := confidence
		labels = append(labels, rekognitiontypes.Label{Name: &value, Confidence: &conf})
	}
	return labels
}
