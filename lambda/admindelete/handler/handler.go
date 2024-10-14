package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/benjaminkitson/bk-auth-api/utils/auth"
	utils "github.com/benjaminkitson/bk-auth-api/utils/lambda"
	"go.uber.org/zap"
)

type UserAPIClient interface {
	DeleteUser(ctx context.Context, id string) (string, error)
}

type handler struct {
	delete        auth.AdapterHandler
	logger        *zap.Logger
	userAPIClient UserAPIClient
}

func NewHandler(logger *zap.Logger, d auth.AdapterHandler, c UserAPIClient) (handler, error) {
	return handler{
		delete:        d,
		logger:        logger,
		userAPIClient: c,
	}, nil
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
	_, err = handler.delete(bodyMap)
	if err != nil {
		handler.logger.Error("Error deleting user from auth provider by email", zap.Error(err))
		return utils.RESPONSE_500, nil
	}

	id, err := handler.userAPIClient.DeleteUser(context.Background(), bodyMap["id"])
	if err != nil {
		// TODO: Error response doesn't work as expected here
		handler.logger.Error("Error deleting user record from db", zap.Error(err))
		return utils.RESPONSE_500, nil
	}
	rm := map[string]string{"id": id}

	r, err := json.Marshal(rm)
	if err != nil {
		handler.logger.Error("admin delete error", zap.Error(err))
		return utils.RESPONSE_500, nil
	}

	return utils.RESPONSE_200(string(r)), nil
}
