package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

type MockAdapter struct {
	isError bool
}

func (ma MockAdapter) SignIn(body map[string]string) (map[string]string, error) {
	if ma.isError {
		return nil, fmt.Errorf("Auth provider error")
	}
	return map[string]string{
		"message": "Successfully signed in!",
	}, nil
}

func (ma MockAdapter) SignUp(body map[string]string) (map[string]string, error) {
	if ma.isError {
		return nil, fmt.Errorf("Auth provider error")
	}
	return map[string]string{
		"message": "Successfully signed up!",
	}, nil
}

func (ma MockAdapter) VerifyEmail(body map[string]string) (map[string]string, error) {
	if ma.isError {
		return nil, fmt.Errorf("Auth provider error")
	}
	return map[string]string{
		"message": "Successfully verified email!",
	}, nil
}

/*
Tests the basic workings of the handler, using a mocked auth provider client that either succeeds or returns some generic error
*/
func TestHandler(t *testing.T) {
	type test struct {
		Name                   string
		AdapterError           bool
		SecretsGetterError     bool
		RequestBody            string
		RequestPath            string
		ExpectedStatusCode     int
		IsHandlerErrorExpected bool
	}

	tests := []test{
		{
			Name:               "Sign in success",
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"password\": \"password\"}",
			RequestPath:        "/signin",
			ExpectedStatusCode: 200,
		},
		{
			Name:               "Sign in auth provider adapter error",
			AdapterError:       true,
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"password\": \"password\"}",
			RequestPath:        "/signin",
			ExpectedStatusCode: 500,
		},
		{
			Name:               "Sign up success",
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"password\": \"abcabc123\"}",
			RequestPath:        "/signup",
			ExpectedStatusCode: 200,
		},
		{
			Name:               "Sign up auth provider adapter error",
			AdapterError:       true,
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"password\": \"password\"}",
			RequestPath:        "/signup",
			ExpectedStatusCode: 500,
		},
		{
			Name:               "Verify email success",
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"code\": \"1234\"}",
			RequestPath:        "/verify",
			ExpectedStatusCode: 200,
		},
		{
			Name:               "Verify email auth provider adapter error",
			AdapterError:       true,
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"code\": \"1234\"}",
			RequestPath:        "/verify",
			ExpectedStatusCode: 500,
		},
		{
			Name:                   "Invalid path supplied",
			RequestBody:            "{\"email\": \"abc@gmail.com\", \"password\": \"password\"}",
			RequestPath:            "/someInvalidPath",
			ExpectedStatusCode:     500,
			IsHandlerErrorExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			l, err := zap.NewDevelopment()
			if err != nil {
				t.Fatalf("Failed to initialise dev logger")
			}

			m := MockAdapter{
				isError: tt.AdapterError,
			}

			h, err := NewHandler(l, m, "someurl.com")
			if err != nil {
				t.Fatalf("Failed to initialise handler")
			}

			req := events.APIGatewayProxyRequest{
				// This test should probably fail if the body isn't the correct format?
				Body: tt.RequestBody,
				Path: tt.RequestPath,
			}

			r, err := h.Handle(context.TODO(), req)
			if err != nil && !tt.IsHandlerErrorExpected {
				t.Fatalf("Unexpected handler error")
			}

			t.Log(r.StatusCode)

			if r.StatusCode != tt.ExpectedStatusCode {
				t.Fatalf("Expected Status Code to be %v", tt.ExpectedStatusCode)
			}
		})
	}
}
