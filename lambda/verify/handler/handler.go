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
	userAPIClient       UserAPIClient
}

func NewHandler(logger *zap.Logger, a AuthProviderAdapter, c UserAPIClient) (handler, error) {
	return handler{
		authProviderAdapter: a,
		logger:              logger,
		userAPIClient:       c,
	}, nil
}

type AdapterHandler func(map[string]string) (map[string]string, error)

type AuthProviderAdapter interface {
	VerifyEmail(map[string]string) (map[string]string, error)
}

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

	// TODO: Might want to return more detailed information when these things go wrong? Maybe in some cases
	// For now we don't actually use the response
	_, err = handler.authProviderAdapter.VerifyEmail(bodyMap)
	if err != nil {
		handler.logger.Error("Error verifying email", zap.Error(err))
		return utils.RESPONSE_500, nil
	}

	u, err := handler.userAPIClient.CreateUser(context.Background(), bodyMap["email"])
	if err != nil {
		handler.logger.Error("Error creating user", zap.Error(err))
		return utils.RESPONSE_500, nil
	}
	rm := map[string]string{"email": u.Email}

	r, err := json.Marshal(rm)
	if err != nil {
		handler.logger.Error("verify email error", zap.Error(err))
		return utils.RESPONSE_500, nil
	}

	return utils.RESPONSE_200(string(r)), nil
}
