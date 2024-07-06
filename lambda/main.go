package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/auth-api/lambda/auth"
	"github.com/auth-api/lambda/secrets"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"go.uber.org/zap"
)

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return handleRequest(ctx, request)
	})
}

// TODO: make distinction between 400 and 500 errors

// Wrapper around json.Marshal that handles generic error cases
func getResponseBody(data any) string {
	e := "{\"message\": \"Something went wrong!\"}"
	if _, ok := data.(error); ok {
		// If the input to getResponseBody is an error, return the generic error message
		return e
	}
	r, err := json.Marshal(data)
	if err != nil {
		// Fallback JSON string in case marshalling somehow goes wrong
		return e
	}
	return string(r)
}

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

	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Error("Failed to intialise SDK config", zap.Error(err))
		s := getResponseBody(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       s,
		}, nil
	}

	cognitoClient := cognitoidentityprovider.NewFromConfig(sdkConfig)

	secretsClient, err := secrets.NewSecretsClient(logger)
	if err != nil {
		logger.Error("failed to initialise secrets client", zap.Error(err))
		s := getResponseBody(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       s,
		}, nil
	}

	// Each of these should be converted into some generalised method that accepts the "auth.SomeFunction" function as an argument
	if request.Path == "/signup" {
		r, err := auth.HandleSignUp(request.Body, secretsClient, cognitoClient, logger)
		if err != nil {
			logger.Error("signup error", zap.Error(err))
			s := getResponseBody(err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers:    headers,
				Body:       s,
			}, nil
		}
		s := getResponseBody(r)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       s,
		}, nil
	}

	if request.Path == "/signin" {
		r, err := auth.HandleSignIn(request.Body, secretsClient, cognitoClient, logger)
		if err != nil {
			logger.Error("signin error", zap.Error(err))
			s := getResponseBody(err)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers:    headers,
				Body:       s,
			}, nil
		}
		s := getResponseBody(r)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       s,
		}, nil
	}

	if request.Path == "/verify" {
		r, err := auth.HandleVerify(request.Body, secretsClient, cognitoClient, logger)
		if err != nil {
			logger.Error("verify error", zap.Error(err))
			s := getResponseBody(r)
			return events.APIGatewayProxyResponse{
				StatusCode: 500,
				Headers:    headers,
				Body:       s,
			}, nil
		}
		s := getResponseBody(r)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       s,
		}, nil
	}

	logger.Error("invalid path")
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Headers:    headers,
	}, fmt.Errorf("invalid path")
}
