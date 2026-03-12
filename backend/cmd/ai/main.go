package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fiando/propti/backend/internal/handlers"
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

	var aiSvc *services.AIService
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		aiSvc = services.NewAIService(key)
	}

	s3Svc, err := services.NewS3Service(ctx, os.Getenv("S3_MEDIA_BUCKET"))
	if err != nil {
		utils.LogError("init s3 service", err)
		panic(err)
	}

	mapsSvc := services.NewGoogleMapsServiceFromEnv()
	listingSvc := services.NewListingService(listingRepo, userRepo, aiSvc, s3Svc, mapsSvc, nil)

	premiumHandler := handlers.NewPremiumHandler(userRepo, listingRepo, transactionRepo, listingSvc)
	lambda.Start(premiumHandler.Handle)
}
