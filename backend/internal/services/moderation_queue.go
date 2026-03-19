package services

import (
	"context"
	"encoding/json"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	lambdasvc "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

const ListingModerationAction = "moderate_listing"

type ListingModerationEvent struct {
	Action      string    `json:"action"`
	ListingID   string    `json:"listingId"`
	CheckText   *bool     `json:"checkText,omitempty"`   // nil = check (backward compat default)
	NewImageIDs *[]string `json:"newImageIds,omitempty"` // nil = check all; empty slice = skip images
}

type lambdaInvoker interface {
	Invoke(ctx context.Context, params *lambdasvc.InvokeInput, optFns ...func(*lambdasvc.Options)) (*lambdasvc.InvokeOutput, error)
}

type LambdaModerationEnqueuer struct {
	client       lambdaInvoker
	functionName string
}

func NewLambdaModerationEnqueuer(ctx context.Context, functionName string) (*LambdaModerationEnqueuer, error) {
	if functionName == "" {
		return nil, nil
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config for moderation queue: %w", err)
	}

	return &LambdaModerationEnqueuer{
		client:       lambdasvc.NewFromConfig(cfg),
		functionName: functionName,
	}, nil
}

func (q *LambdaModerationEnqueuer) EnqueueListingModeration(ctx context.Context, listingID string, checkText bool, newImageIDs []string) error {
	if q == nil || q.functionName == "" {
		return nil
	}

	ct := checkText
	ni := newImageIDs
	payload, err := json.Marshal(ListingModerationEvent{
		Action:      ListingModerationAction,
		ListingID:   listingID,
		CheckText:   &ct,
		NewImageIDs: &ni,
	})
	if err != nil {
		return fmt.Errorf("marshal moderation event: %w", err)
	}

	out, err := q.client.Invoke(ctx, &lambdasvc.InvokeInput{
		FunctionName:   &q.functionName,
		InvocationType: lambdatypes.InvocationTypeEvent,
		Payload:        payload,
	})
	if err != nil {
		return fmt.Errorf("invoke moderation worker: %w", err)
	}
	if out.StatusCode < 200 || out.StatusCode >= 300 {
		return fmt.Errorf("invoke moderation worker returned status %d", out.StatusCode)
	}

	return nil
}
