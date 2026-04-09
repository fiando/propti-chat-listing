package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type OTPChallenge struct {
	ChallengeID  string     `json:"challengeId" dynamodbav:"otpId"`
	UserID       string     `json:"userId" dynamodbav:"userId"`
	Phone        string     `json:"phone" dynamodbav:"phone"`
	OTPCode      string     `json:"-" dynamodbav:"otpCode"`
	RetryCount   int        `json:"retryCount" dynamodbav:"retryCount"`
	AttemptCount int        `json:"attemptCount" dynamodbav:"attemptCount"`
	MaxAttempts  int        `json:"maxAttempts" dynamodbav:"maxAttempts"`
	ExpiresAt    time.Time  `json:"expiresAt" dynamodbav:"expiresAt"`
	VerifiedAt   *time.Time `json:"verifiedAt,omitempty" dynamodbav:"verifiedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt" dynamodbav:"updatedAt"`
}

type OTPStore interface {
	Put(ctx context.Context, challenge *OTPChallenge) error
	GetByID(ctx context.Context, challengeID string) (*OTPChallenge, error)
	GetLatestByUser(ctx context.Context, userID string) (*OTPChallenge, error)
	GetLatestByPhone(ctx context.Context, phone string) (*OTPChallenge, error)
}

type OTPRepo struct {
	db *DynamoDB
}

func NewOTPRepo(db *DynamoDB) *OTPRepo {
	return &OTPRepo{db: db}
}

func (r *OTPRepo) Put(ctx context.Context, challenge *OTPChallenge) error {
	if challenge == nil {
		return fmt.Errorf("challenge is required")
	}
	if challenge.ChallengeID == "" {
		return fmt.Errorf("challenge id is required")
	}

	item, err := attributevalue.MarshalMap(challenge)
	if err != nil {
		return fmt.Errorf("marshal otp challenge: %w", err)
	}

	_, err = r.db.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.db.OTPChallengesTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put otp challenge: %w", err)
	}
	return nil
}

func (r *OTPRepo) GetByID(ctx context.Context, challengeID string) (*OTPChallenge, error) {
	if challengeID == "" {
		return nil, fmt.Errorf("challenge id is required")
	}

	result, err := r.db.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.db.OTPChallengesTable),
		Key: map[string]types.AttributeValue{
			"otpId": &types.AttributeValueMemberS{Value: challengeID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get otp challenge: %w", err)
	}
	if len(result.Item) == 0 {
		return nil, nil
	}

	var challenge OTPChallenge
	if err := attributevalue.UnmarshalMap(result.Item, &challenge); err != nil {
		return nil, fmt.Errorf("unmarshal otp challenge: %w", err)
	}
	return &challenge, nil
}

func (r *OTPRepo) GetLatestByUser(ctx context.Context, userID string) (*OTPChallenge, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	result, err := r.db.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.db.OTPChallengesTable),
		FilterExpression: aws.String("userId = :uid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":uid": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scan otp challenges by user: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	challenges := make([]OTPChallenge, 0, len(result.Items))
	for _, item := range result.Items {
		var challenge OTPChallenge
		if err := attributevalue.UnmarshalMap(item, &challenge); err != nil {
			return nil, fmt.Errorf("unmarshal otp challenge: %w", err)
		}
		challenges = append(challenges, challenge)
	}

	sort.Slice(challenges, func(i, j int) bool {
		return challenges[i].CreatedAt.After(challenges[j].CreatedAt)
	})

	latest := challenges[0]
	return &latest, nil
}

func (r *OTPRepo) GetLatestByPhone(ctx context.Context, phone string) (*OTPChallenge, error) {
	if phone == "" {
		return nil, fmt.Errorf("phone is required")
	}

	result, err := r.db.Client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.db.OTPChallengesTable),
		FilterExpression: aws.String("phone = :phone"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":phone": &types.AttributeValueMemberS{Value: phone},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scan otp challenges by phone: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	challenges := make([]OTPChallenge, 0, len(result.Items))
	for _, item := range result.Items {
		var challenge OTPChallenge
		if err := attributevalue.UnmarshalMap(item, &challenge); err != nil {
			return nil, fmt.Errorf("unmarshal otp challenge: %w", err)
		}
		challenges = append(challenges, challenge)
	}

	sort.Slice(challenges, func(i, j int) bool {
		return challenges[i].CreatedAt.After(challenges[j].CreatedAt)
	})

	latest := challenges[0]
	return &latest, nil
}
