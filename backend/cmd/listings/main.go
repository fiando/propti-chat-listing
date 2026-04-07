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
	leadRepo := repository.NewLeadRepo(db)
	userRepo := repository.NewUserRepo(db)
	moderationRepo := repository.NewModerationRepo(db)
	uploadSessionRepo := repository.NewUploadSessionRepo(db)

	s3Svc, err := services.NewS3Service(ctx, os.Getenv("S3_MEDIA_BUCKET"))
	if err != nil {
		utils.LogError("init s3 service", err)
		panic(err)
	}

	var aiSvc *services.AIService
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	if openAIAPIKey != "" {
		aiSvc = services.NewAIService(openAIAPIKey)
	}

	mapsSvc := services.NewGoogleMapsServiceFromEnv()

	locationCatalog, err := services.NewLocationCatalogFromReader(bytes.NewReader(data.IndonesiaLocationsJSON))
	if err != nil {
		utils.LogError("init location catalog", err)
		panic(err)
	}
	listingSvc := services.NewListingService(listingRepo, userRepo, aiSvc, s3Svc, mapsSvc, locationCatalog)
	leadSvc := services.NewLeadService(leadRepo)
	listingSvc.SetUploadSessionStore(uploadSessionRepo)
	uploadSessionSvc := services.NewUploadSessionService(uploadSessionRepo, userRepo, listingRepo, s3Svc)
	searchIntentSvc := services.NewSearchIntentService(aiSvc, locationCatalog)
	moderationQueue, err := services.NewLambdaModerationEnqueuer(ctx, os.Getenv("LISTINGS_FUNCTION_NAME"))
	if err != nil {
		utils.LogError("init moderation queue", err)
		panic(err)
	}
	if moderationQueue != nil {
		listingSvc.SetModerationEnqueuer(moderationQueue)
	}

	imageModerator, err := services.NewImageModerator(ctx, os.Getenv("IMAGE_MODERATION_PROVIDER"), openAIAPIKey)
	if err != nil {
		utils.LogError("init image moderator", err)
		panic(err)
	}

	var textModerator services.ContentModerator
	if aiSvc != nil {
		textModerator = aiSvc
	}
	moderationSvc := services.NewModerationService(textModerator, imageModerator, moderationRepo, listingRepo, s3Svc)

	listingHandler := handlers.NewListingHandler(listingSvc, uploadSessionSvc, services.NewListingMediaPresenter(s3Svc))
	searchHandler := handlers.NewSearchHandler(listingRepo, mapsSvc, locationCatalog, searchIntentSvc)
	leadHandler := handlers.NewLeadHandler(leadSvc)

	// Route /search/* and /locations/* to searchHandler; all others to listingHandler.
	lambda.Start(func(ctx context.Context, rawReq json.RawMessage) (interface{}, error) {
		var moderationEvent services.ListingModerationEvent
		if err := json.Unmarshal(rawReq, &moderationEvent); err == nil && moderationEvent.Action == services.ListingModerationAction {
			// Default to checking everything for backward compatibility with older events
			checkText := moderationEvent.CheckText == nil || *moderationEvent.CheckText
			var newImageIDs []string
			if moderationEvent.NewImageIDs != nil {
				newImageIDs = *moderationEvent.NewImageIDs
			}
			listing, err := moderationSvc.ModerateListing(ctx, moderationEvent.ListingID, checkText, newImageIDs)
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

		return handlers.CombinedListingHandler(ctx, req, listingHandler, searchHandler, leadHandler)
	})
}
