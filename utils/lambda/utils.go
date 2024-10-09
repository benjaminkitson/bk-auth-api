package utils

// TODO: Not sure if this package name follows correct convention with respect to folder structure

import "github.com/aws/aws-lambda-go/events"

var Headers = map[string]string{
	"Access-Control-Allow-Headers": "Content-Type",
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "OPTIONS,POST,GET",
}

var RESPONSE_500 = events.APIGatewayProxyResponse{
	StatusCode: 500,
	Headers:    Headers,
	Body:       "{\"message\": \"Something went wrong!\"}",
}

var RESPONSE_400 = events.APIGatewayProxyResponse{
	StatusCode: 400,
	Headers:    Headers,
	Body:       "{\"message\": \"Invalid request\"}",
}

func RESPONSE_200(body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    Headers,
		Body:       body,
	}
}
