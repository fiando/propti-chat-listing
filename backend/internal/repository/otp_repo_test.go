package repository

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

func TestOTPChallengeMarshal_UsesOtpIDKey(t *testing.T) {
	challenge := OTPChallenge{
		ChallengeID: "challenge-123",
		UserID:      "user-1",
		Phone:       "6281234567890",
		OTPCode:     "123456",
		ExpiresAt:   time.Now().UTC().Add(5 * time.Minute),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	item, err := attributevalue.MarshalMap(challenge)
	if err != nil {
		t.Fatalf("marshal challenge: %v", err)
	}

	if _, ok := item["otpId"]; !ok {
		t.Fatalf("expected marshaled item to contain otpId")
	}
	if _, ok := item["challengeId"]; ok {
		t.Fatalf("expected marshaled item to not contain legacy challengeId key")
	}
}
