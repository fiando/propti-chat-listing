package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/fiando/propti/backend/internal/handlers"
	"github.com/fiando/propti/backend/internal/repository"
	"github.com/fiando/propti/backend/internal/utils"
)

func main() {
	ctx := context.Background()

	db, err := repository.NewDynamoDB(ctx)
	if err != nil {
		utils.LogError("init dynamodb", err)
		panic(err)
	}

	handler := handlers.NewAuthHandler(db)
	lambda.Start(handler.Handle)
}
