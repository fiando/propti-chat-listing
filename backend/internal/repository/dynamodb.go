package repository

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// DynamoDB holds a DynamoDB client and the resolved table names.
type DynamoDB struct {
	Client                *dynamodb.Client
	ListingsTable         string
	UsersTable            string
	TransactionsTable     string
	ModerationsTable      string
	LeadsTable            string
	UploadSessionsTable   string
	WhatsAppSessionsTable string
	OTPChallengesTable    string
}

type dynamoDBClientConfig struct {
	Region          string
	EndpointURL     string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// NewDynamoDB creates a DynamoDB helper using the ambient AWS config.
func NewDynamoDB(ctx context.Context) (*DynamoDB, error) {
	clientConfig := resolveDynamoDBClientConfigFromEnv()

	loadOptions := []func(*config.LoadOptions) error{}
	if clientConfig.Region != "" {
		loadOptions = append(loadOptions, config.WithRegion(clientConfig.Region))
	}
	if clientConfig.AccessKeyID != "" || clientConfig.SecretAccessKey != "" || clientConfig.SessionToken != "" {
		loadOptions = append(
			loadOptions,
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					clientConfig.AccessKeyID,
					clientConfig.SecretAccessKey,
					clientConfig.SessionToken,
				),
			),
		)
	}

	cfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)
	if clientConfig.EndpointURL != "" {
		client = dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
			options.BaseEndpoint = aws.String(clientConfig.EndpointURL)
		})
	}

	return &DynamoDB{
		Client:                client,
		ListingsTable:         getEnv("DYNAMODB_LISTINGS_TABLE", "propti-listings"),
		UsersTable:            getEnv("DYNAMODB_USERS_TABLE", "propti-users"),
		TransactionsTable:     getEnv("DYNAMODB_TRANSACTIONS_TABLE", "propti-transactions"),
		ModerationsTable:      getEnv("DYNAMODB_MODERATIONS_TABLE", "propti-moderations"),
		LeadsTable:            getEnv("DYNAMODB_LEADS_TABLE", "propti-leads"),
		UploadSessionsTable:   getEnv("DYNAMODB_UPLOAD_SESSIONS_TABLE", "propti-upload-sessions"),
		WhatsAppSessionsTable: getEnv("DYNAMODB_WHATSAPP_SESSIONS_TABLE", "propti-whatsapp-sessions"),
		OTPChallengesTable:    resolveOTPChallengesTableEnv(),
	}, nil
}

func resolveDynamoDBClientConfigFromEnv() dynamoDBClientConfig {
	endpointURL := os.Getenv("DYNAMODB_ENDPOINT_URL")
	if endpointURL == "" {
		return dynamoDBClientConfig{}
	}

	region := getEnv("AWS_REGION", "ap-southeast-1")
	accessKeyID := getEnv("AWS_ACCESS_KEY_ID", "local")
	secretAccessKey := getEnv("AWS_SECRET_ACCESS_KEY", "local")

	return dynamoDBClientConfig{
		Region:          region,
		EndpointURL:     endpointURL,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func resolveOTPChallengesTableEnv() string {
	if val := os.Getenv("DYNAMODB_OTP_CHALLENGES_TABLE"); val != "" {
		return val
	}
	if val := os.Getenv("DYNAMODB_WHATSAPP_OTP_TABLE"); val != "" {
		return val
	}
	return "propti-otp-challenges"
}
