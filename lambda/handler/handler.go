package handler

import (
	"context"
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

type AdapterHandler func(string) (string, error)

type AuthProviderAdapter interface {
	SignIn(string) (string, error)
	SignUp(string) (string, error)
	VerifyEmail(string) (string, error)
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
	handler.logger.Info("request", zap.Any("request", request))

	// TODO: this is clever but too confusing
	if request.Path == "/signup" {
		return handler.handlePath(handler.authProviderAdapter.SignUp, request.Body)
	}

	if request.Path == "/signin" {
		return handler.handlePath(handler.authProviderAdapter.SignIn, request.Body)
	}

	if request.Path == "/verify" {
		return handler.handlePath(handler.authProviderAdapter.VerifyEmail, request.Body)
	}

	handler.logger.Error("invalid path")
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Headers:    Headers,
		// Error body needed? Probably not
	}, fmt.Errorf("invalid path")
}

func (handler handler) handlePath(pathFunc AdapterHandler, rb string) (events.APIGatewayProxyResponse, error) {
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
