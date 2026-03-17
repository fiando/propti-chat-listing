package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fiando/propti/backend/internal/data"
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
	moderationRepo := repository.NewModerationRepo(db)

	s3Svc, err := services.NewS3Service(ctx, os.Getenv("S3_MEDIA_BUCKET"))
	if err != nil {
		utils.LogError("init s3 service", err)
		panic(err)
	}

	var aiSvc services.AIParseService
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		aiSvc = services.NewAIService(key)
	}

	mapsSvc := services.NewGoogleMapsServiceFromEnv()

	locationCatalog, err := services.NewLocationCatalogFromReader(bytes.NewReader(data.IndonesiaLocationsJSON))
	if err != nil {
		utils.LogError("init location catalog", err)
		panic(err)
	}
	listingSvc := services.NewListingService(listingRepo, userRepo, aiSvc, s3Svc, mapsSvc, locationCatalog)
	moderationQueue, err := services.NewLambdaModerationEnqueuer(ctx, os.Getenv("LISTINGS_FUNCTION_NAME"))
	if err != nil {
		utils.LogError("init moderation queue", err)
		panic(err)
	}
	if moderationQueue != nil {
		listingSvc.SetModerationEnqueuer(moderationQueue)
	}

	imageModerator, err := services.NewRekognitionImageModerator(ctx)
	if err != nil {
		utils.LogError("init image moderator", err)
		panic(err)
	}

	var textModerator services.ContentModerator
	if moderator, ok := aiSvc.(services.ContentModerator); ok {
		textModerator = moderator
	}
	moderationSvc := services.NewModerationService(textModerator, imageModerator, moderationRepo, listingRepo)

	listingHandler := handlers.NewListingHandler(listingSvc, userRepo)
	searchHandler := handlers.NewSearchHandler(listingRepo, mapsSvc, locationCatalog)

	// Route /search/* and /locations/* to searchHandler; all others to listingHandler.
	lambda.Start(func(ctx context.Context, rawReq json.RawMessage) (interface{}, error) {
		var moderationEvent services.ListingModerationEvent
		if err := json.Unmarshal(rawReq, &moderationEvent); err == nil && moderationEvent.Action == services.ListingModerationAction {
			listing, err := moderationSvc.ModerateListing(ctx, moderationEvent.ListingID)
			if err != nil {
				utils.LogError("moderate listing asynchronously", err, "listingId", moderationEvent.ListingID)
				return map[string]string{
					"status": "error",
				}, err
			}

			return map[string]string{
				"status":    "processed",
				"listingId": listing.ListingID,
			}, nil
		}

		var req interface{}
		if err := json.Unmarshal(rawReq, &req); err != nil {
			utils.LogError("decode lambda request", err)
			return nil, err
		}

		return handlers.CombinedListingHandler(ctx, req, listingHandler, searchHandler)
	})
}
