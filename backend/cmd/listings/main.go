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

	listingRepo := repository.NewListingRepo(db)
	userRepo := repository.NewUserRepo(db)

	s3Svc, err := services.NewS3Service(ctx, os.Getenv("S3_MEDIA_BUCKET"))
	if err != nil {
		utils.LogError("init s3 service", err)
		panic(err)
	}

	var aiSvc *services.AIService
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		aiSvc = services.NewAIService(key)
	}

	mapsSvc := services.NewGoogleMapsServiceFromEnv()

	listingSvc := services.NewListingService(listingRepo, userRepo, aiSvc, s3Svc, mapsSvc)

	listingHandler := handlers.NewListingHandler(listingSvc, userRepo)
	searchHandler := handlers.NewSearchHandler(listingRepo, mapsSvc)

	// Route /search/* and /locations/* to searchHandler; all others to listingHandler.
	lambda.Start(func(ctx context.Context, req interface{}) (interface{}, error) {
		// Use a combined handler that dispatches based on path prefix.
		return handlers.CombinedListingHandler(ctx, req, listingHandler, searchHandler)
	})
}
