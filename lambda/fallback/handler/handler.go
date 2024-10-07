package handler

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

type handler struct {
	logger *zap.Logger
}

func NewHandler(logger *zap.Logger) (handler, error) {
	return handler{
		logger: logger,
	}, nil
}

var Headers = map[string]string{
	"Access-Control-Allow-Headers": "Content-Type",
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "OPTIONS,POST,GET",
}

func (handler handler) Handle(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	handler.logger.Error("invalid path")
	return events.APIGatewayProxyResponse{
		StatusCode: 400,
		Headers:    Headers,
		Body:       "{\"message\": \"Invalid path\"}",
	}, nil
}
