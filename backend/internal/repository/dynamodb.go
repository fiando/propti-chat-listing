package repository

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDB holds a DynamoDB client and the resolved table names.
type DynamoDB struct {
	Client                *dynamodb.Client
	ListingsTable         string
	UsersTable            string
	TransactionsTable     string
	ModerationsTable      string
	UploadSessionsTable   string
	WhatsAppSessionsTable string
	OTPChallengesTable    string
}

// NewDynamoDB creates a DynamoDB helper using the ambient AWS config.
func NewDynamoDB(ctx context.Context) (*DynamoDB, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &DynamoDB{
		Client:                dynamodb.NewFromConfig(cfg),
		ListingsTable:         getEnv("DYNAMODB_LISTINGS_TABLE", "propti-listings"),
		UsersTable:            getEnv("DYNAMODB_USERS_TABLE", "propti-users"),
		TransactionsTable:     getEnv("DYNAMODB_TRANSACTIONS_TABLE", "propti-transactions"),
		ModerationsTable:      getEnv("DYNAMODB_MODERATIONS_TABLE", "propti-moderations"),
		UploadSessionsTable:   getEnv("DYNAMODB_UPLOAD_SESSIONS_TABLE", "propti-upload-sessions"),
		WhatsAppSessionsTable: getEnv("DYNAMODB_WHATSAPP_SESSIONS_TABLE", "propti-whatsapp-sessions"),
		OTPChallengesTable:    getEnv("DYNAMODB_OTP_CHALLENGES_TABLE", "propti-otp-challenges"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
