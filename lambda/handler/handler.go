package handler

import (
	"context"
	"fmt"

	"github.com/auth-api/lambda/cognito-adapter"
	"github.com/auth-api/lambda/secrets"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

type handler struct {
	headers        map[string]string
	secretsClient  secrets.SecretGetter
	cognitoAdapter cognito.Adapter
	logger         *zap.Logger
	genericError   string
}

func NewHandler(logger *zap.Logger, sc secrets.SecretGetter, ca cognito.Adapter) (handler, error) {
	h := map[string]string{
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "OPTIONS,POST,GET",
	}

	e := "{\"message\": \"Something went wrong!\"}"

	return handler{
		headers:        h,
		secretsClient:  sc,
		cognitoAdapter: ca,
		logger:         logger,
		genericError:   e,
	}, nil
}

// TODO: make distinction between 400 and 500 errors

func (handler handler) Handle(_ context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	handler.logger.Info("request", zap.Any("request", request))

	// Each of these should be converted into some generalised method that accepts the "auth.SomeFunction" function as an argument
	if request.Path == "/signup" {
		return handler.handlePath(handler.cognitoAdapter.SignUp, request.Body)
	}

	if request.Path == "/signin" {
		return handler.handlePath(handler.cognitoAdapter.SignIn, request.Body)
	}

	if request.Path == "/verify" {
		return handler.handlePath(handler.cognitoAdapter.VerifyEmail, request.Body)
	}

	handler.logger.Error("invalid path")
	return events.APIGatewayProxyResponse{
		StatusCode: 500,
		Headers:    handler.headers,
	}, fmt.Errorf("invalid path")
}

func (handler handler) handlePath(pathFunc cognito.AdapterHandler, rb string) (events.APIGatewayProxyResponse, error) {
	r, err := pathFunc(rb)
	if err != nil {
		handler.logger.Error("signin error", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    handler.headers,
			Body:       handler.genericError,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    handler.headers,
		Body:       r,
	}, nil
}
