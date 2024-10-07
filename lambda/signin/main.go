package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	cognito "github.com/benjaminkitson/bk-auth-api/cognitoadapter"
	"github.com/benjaminkitson/bk-auth-api/lambda/signin/handler"
	"github.com/benjaminkitson/bk-auth-api/secrets"
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

		sdkConfig, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			logger.Error("Failed to intialise SDK config", zap.Error(err))
			return events.APIGatewayProxyResponse{}, err
		}

		sm := secretsmanager.NewFromConfig(sdkConfig)
		sc, err := secrets.NewSecretsClient(logger, sm)
		if err != nil {
			logger.Error("failed to initialise secrets client", zap.Error(err))
			return events.APIGatewayProxyResponse{}, err
		}

		cc := cognitoidentityprovider.NewFromConfig(sdkConfig)
		// TODO: investigate just storing this as an env var from the CDK
		ccid, err := sc.GetSecret("COGNITO_CLIENT")
		if err != nil {
			logger.Error("Failed to get cognito client id", zap.Error(err))
			return events.APIGatewayProxyResponse{}, err
		}

		ca := cognito.NewAdapter(cc, ccid, logger)

		h, err := handler.NewHandler(logger, ca)
		if err != nil {
			return events.APIGatewayProxyResponse{}, err
		}

		return h.Handle(ctx, request)
	})
}
