package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

type handler struct {
	authProviderAdapter AuthProviderAdapter
	logger              *zap.Logger
}

func NewHandler(logger *zap.Logger, a AuthProviderAdapter) (handler, error) {
	return handler{
		authProviderAdapter: a,
		logger:              logger,
	}, nil
}

type AdapterHandler func(map[string]string) (string, error)

type AuthProviderAdapter interface {
	SignIn(map[string]string) (string, error)
	SignUp(map[string]string) (string, error)
	VerifyEmail(map[string]string) (string, error)
}

var Headers = map[string]string{
	"Access-Control-Allow-Headers": "Content-Type",
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "OPTIONS,POST,GET",
}

const GenericError = "{\"message\": \"Something went wrong!\"}"

// TODO: make distinction between 400 and 500 errors
// TODO: understand how different methods are dealt with (post vs get etc)
// TODO: probably incorporate some sort of request body validation prior to calling cognito or whichever auth provider

func (handler handler) Handle(_ context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bodyMap := make(map[string]string)

	err := json.Unmarshal([]byte(request.Body), &bodyMap)
	if err != nil {
		handler.logger.Error("Error parsing request body", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    Headers,
			// Error body needed? Probably not
		}, fmt.Errorf("error parsing request body")
	}

	// TODO: this is clever but too confusing
	if request.Path == "/signup" {
		return handler.handlePath(handler.authProviderAdapter.SignUp, bodyMap)
	}

	if request.Path == "/signin" {
		return handler.handlePath(handler.authProviderAdapter.SignIn, bodyMap)
	}

	if request.Path == "/verify" {
		return handler.handlePath(handler.authProviderAdapter.VerifyEmail, bodyMap)
	}

	handler.logger.Error("invalid path")
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Headers:    Headers,
		// Error body needed? Probably not
	}, fmt.Errorf("invalid path")
}

func (handler handler) handlePath(pathFunc AdapterHandler, rb map[string]string) (events.APIGatewayProxyResponse, error) {
	r, err := pathFunc(rb)
	if err != nil {
		handler.logger.Error("signin error", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    Headers,
			Body:       GenericError,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    Headers,
		Body:       r,
	}, nil
}
