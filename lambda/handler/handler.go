package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/benjaminkitson/bk-user-api/client"
	"go.uber.org/zap"
)

type handler struct {
	authProviderAdapter AuthProviderAdapter
	logger              *zap.Logger
	apiURL              string
}

func NewHandler(logger *zap.Logger, a AuthProviderAdapter, u string) (handler, error) {
	return handler{
		authProviderAdapter: a,
		logger:              logger,
		apiURL:              u,
	}, nil
}

type AdapterHandler func(map[string]string) (map[string]string, error)

type AuthProviderAdapter interface {
	SignIn(map[string]string) (map[string]string, error)
	SignUp(map[string]string) (map[string]string, error)
	VerifyEmail(map[string]string) (map[string]string, error)
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
		return handler.handlePath(func(rb map[string]string) (map[string]string, error) {
			// TODO: This should really happen at the verification stage, not the initial sign up stage
			b, err := handler.authProviderAdapter.SignUp(rb)
			if err != nil {
				return nil, err
			}
			_, err = json.Marshal(b)
			if err != nil {
				return nil, err
			}
			c, err := client.NewClient("https://api.benjaminkitson.com", handler.logger)
			if err != nil {
				return nil, err
			}
			u, err := c.CreateUser(context.Background(), rb["email"])
			if err != nil {
				return nil, err
			}
			return map[string]string{"email": u.Email}, nil
		}, bodyMap)
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
	d, err := pathFunc(rb)
	if err != nil {
		handler.logger.Error("Failed to get response body from Cognito adapter", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    Headers,
			Body:       GenericError,
		}, nil
	}
	r, err := json.Marshal(d)
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
		Body:       string(r),
	}, nil
}
