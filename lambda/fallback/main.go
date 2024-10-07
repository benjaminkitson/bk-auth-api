package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/benjaminkitson/bk-user-api/lambda/fallback/handler"
	"go.uber.org/zap"
)

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		logger, err := zap.NewProduction()
		if err != nil {
			fmt.Printf("Failed to initialise logger: %v", err)
			logger = &zap.Logger{}
		}
		defer logger.Sync()

		h, err := handler.NewHandler(logger)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		return h.Handle(ctx, request)
	})
}
