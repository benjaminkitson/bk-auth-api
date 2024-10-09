package handler

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	utils "github.com/benjaminkitson/bk-auth-api/utils/lambda"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.Logger
}

// TODO: this can probably be shared as part of some package

func NewHandler(logger *zap.Logger) (handler, error) {
	return handler{
		logger: logger,
	}, nil
}

func (handler handler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	handler.logger.Error("invalid path")
	return utils.RESPONSE_400, nil
}
