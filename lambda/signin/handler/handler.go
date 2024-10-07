package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	utils "github.com/benjaminkitson/bk-auth-api/utils/lambda"
	"github.com/benjaminkitson/bk-user-api/models"
	"go.uber.org/zap"
)

type UserAPIClient interface {
	CreateUser(ctx context.Context, email string) (models.User, error)
}

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

type AdapterHandler func(map[string]string) (map[string]string, error)

type AuthProviderAdapter interface {
	SignIn(map[string]string) (map[string]string, error)
}

// TODO: make distinction between 400 and 500 errors
// TODO: understand how different methods are dealt with (post vs get etc)
// TODO: probably incorporate some sort of request body validation prior to calling cognito or whichever auth provider

func (handler handler) Handle(_ context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: Path check? Not sure if needed
	// handler.logger.Error("invalid path", zap.String("path", request.Path))
	// return utils.RESPONSE_400, nil

	bodyMap := make(map[string]string)

	err := json.Unmarshal([]byte(request.Body), &bodyMap)
	if err != nil {
		handler.logger.Error("Error parsing request body", zap.Error(err))
		return utils.RESPONSE_500, fmt.Errorf("error parsing request body")
	}

	d, err := handler.authProviderAdapter.SignIn(bodyMap)
	if err != nil {
		handler.logger.Error("Failed to get response body from Cognito adapter", zap.Error(err))
		return utils.RESPONSE_500, nil
	}
	r, err := json.Marshal(d)
	if err != nil {
		handler.logger.Error("signin error", zap.Error(err))
		return utils.RESPONSE_500, nil
	}
	return utils.RESPONSE_200(string(r)), nil
}
