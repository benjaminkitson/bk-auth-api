package main

import (
	"context"
	"fmt"

	"github.com/auth-api/lambda/cognito-adapter"
	"github.com/auth-api/lambda/handler"
	"github.com/auth-api/lambda/secrets"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		logger, err := zap.NewProduction()
		if err != nil {
			fmt.Print("Failed to initialise logger")
			logger = &zap.Logger{}
		}
		defer logger.Sync()

		sc, err := secrets.NewSecretsClient(logger)
		if err != nil {
			logger.Error("failed to initialise secrets client", zap.Error(err))
			return events.APIGatewayProxyResponse{}, err
		}

		ca, err := cognito.NewAdapter(sc, logger)
		if err != nil {
			logger.Error("failed to initialise secrets client", zap.Error(err))
			return events.APIGatewayProxyResponse{}, err
		}

		h, err := handler.NewHandler(logger, sc, ca)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		return h.Handle(ctx, request)
	})
}

// TODO: make distinction between 400 and 500 errors

const genericError = "{\"message\": \"Something went wrong!\"}"

/*
Handler function for requests to the auth API - debatable how scalable this approach is, and the code is currently too coupled to Cognito as an auth provider
The auth package in general should probably be converted into a dedicated "Cognito Adapter" with the sign up, sign in etc methods
*/
