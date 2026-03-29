package main

import (
	"bytes"
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fiando/propti/backend/internal/data"
	"github.com/fiando/propti/backend/internal/handlers"
	"github.com/fiando/propti/backend/internal/payments"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/services"
	"github.com/fiando/propti/backend/internal/utils"
)

func main() {
	ctx := context.Background()

	db, err := repository.NewDynamoDB(ctx)
	if err != nil {
		utils.LogError("init dynamodb", err)
		panic(err)
	}

	userRepo := repository.NewUserRepo(db)
	listingRepo := repository.NewListingRepo(db)
	transactionRepo := repository.NewTransactionRepo(db)
	otpRepo := repository.NewOTPRepo(db)

	var aiSvc services.AIParseService
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		aiSvc = services.NewAIService(key)
	}

	s3Svc, err := services.NewS3Service(ctx, os.Getenv("S3_MEDIA_BUCKET"))
	if err != nil {
		utils.LogError("init s3 service", err)
		panic(err)
	}

	mapsSvc := services.NewGoogleMapsServiceFromEnv()

	locationCatalog, err := services.NewLocationCatalogFromReader(bytes.NewReader(data.IndonesiaLocationsJSON))
	if err != nil {
		utils.LogError("init location catalog", err)
		panic(err)
	}

	listingSvc := services.NewListingService(listingRepo, userRepo, aiSvc, s3Svc, mapsSvc, locationCatalog)
	identitySvc, err := services.NewWhatsAppIdentityService(userRepo, otpRepo, services.WhatsAppIdentityOptions{})
	if err != nil {
		utils.LogError("init whatsapp identity service", err)
		panic(err)
	}
	listingSvc.SetWriteEligibilityGuard(identitySvc)
	dokuBaseURL := "https://api-sandbox.doku.com"
	if os.Getenv("DOKU_ENV") == "production" {
		dokuBaseURL = "https://api.doku.com"
	}

	paymentProvider := payments.NewDOKUProvider(payments.DOKUConfig{
		ClientID:  os.Getenv("DOKU_CLIENT_ID"),
		SecretKey: os.Getenv("DOKU_SECRET_KEY"),
		BaseURL:   dokuBaseURL,
	})

	premiumHandler := handlers.NewPremiumHandler(userRepo, transactionRepo, listingSvc, paymentProvider)
	lambda.Start(premiumHandler.Handle)
}
