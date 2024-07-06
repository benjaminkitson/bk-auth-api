package main

import (
	"context"
	"fmt"

	"github.com/auth-api/lambda/cognito-adapter"
	"github.com/auth-api/lambda/secrets"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return handleRequest(ctx, request)
	})
}

// TODO: make distinction between 400 and 500 errors

const genericError = "{\"message\": \"Something went wrong!\"}"

/*
Handler function for requests to the auth API - debatable how scalable this approach is, and the code is currently too coupled to Cognito as an auth provider
The auth package in general should probably be converted into a dedicated "Cognito Adapter" with the sign up, sign in etc methods
*/
func handleRequest(_ context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Print("Failed to initialise logger")
		logger = &zap.Logger{}
	}
	defer logger.Sync()

	logger.Info("request", zap.Any("request", request))

	headers := map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "OPTIONS,POST,GET",
	}

	secretsClient, err := secrets.NewSecretsClient(logger)
	if err != nil {
		logger.Error("failed to initialise secrets client", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       genericError,
		}, nil
	}

	CognitoAdapter, err := cognito.NewAdapter(secretsClient, logger)
	if err != nil {
		logger.Error("failed to initialise secrets client", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       genericError,
		}, nil
	}

	// Each of these should be converted into some generalised method that accepts the "auth.SomeFunction" function as an argument
	if request.Path == "/signup" {
		r, err := CognitoAdapter.SignUp(request.Body)
		if err != nil {
			logger.Error("signup error", zap.Error(err))
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers:    headers,
				Body:       genericError,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       r,
		}, nil
	}

	if request.Path == "/signin" {
		r, err := CognitoAdapter.SignIn(request.Body)
		if err != nil {
			logger.Error("signin error", zap.Error(err))
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers:    headers,
				Body:       genericError,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       r,
		}, nil
	}

	if request.Path == "/verify" {
		r, err := CognitoAdapter.VerifyEmail(request.Body)
		if err != nil {
			logger.Error("verify error", zap.Error(err))
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers:    headers,
				Body:       genericError,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       r,
		}, nil
	}

	logger.Error("invalid path")
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Headers:    headers,
	}, fmt.Errorf("invalid path")
}
