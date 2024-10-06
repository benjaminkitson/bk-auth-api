package handler

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/benjaminkitson/bk-user-api/models"
	"github.com/stretchr/testify/assert"
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

type MockUserAPIClient struct {
	isError bool
}

func (c MockUserAPIClient) CreateUser(ctx context.Context, email string) (models.User, error) {
	if c.isError {
		return models.User{}, fmt.Errorf("API Client Error")
	}
	return models.User{
		Email: email,
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
		Name               string
		AdapterError       bool
		UserAPIClientError bool
		SecretsGetterError bool
		RequestBody        string
		RequestPath        string
		ExpectedStatusCode int
	}

	tests := []test{
		{
			Name:               "Sign in success",
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"password\": \"abcabc123\"}",
			RequestPath:        "/signin",
			ExpectedStatusCode: 200,
		},
		{
			Name:               "Sign in auth provider adapter error",
			AdapterError:       true,
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"password\": \"abcabc123\"}",
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
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"password\": \"abcabc123\"}",
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
			Name:               "Verify email user api client error",
			UserAPIClientError: true,
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"code\": \"1234\"}",
			RequestPath:        "/verify",
			ExpectedStatusCode: 500,
		},
		{
			Name:               "Invalid path supplied",
			RequestBody:        "{\"email\": \"abc@gmail.com\", \"password\": \"password\"}",
			RequestPath:        "/someInvalidPath",
			ExpectedStatusCode: 400,
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

			c := MockUserAPIClient{
				isError: tt.UserAPIClientError,
			}

			h, err := NewHandler(l, m, c)
			assert.Nil(t, err)

			req := events.APIGatewayProxyRequest{
				// This test should probably fail if the body isn't the correct format?
				Body: tt.RequestBody,
				Path: tt.RequestPath,
			}

			r, err := h.Handle(context.TODO(), req)
			assert.Nil(t, err)

			assert.Equal(t, tt.ExpectedStatusCode, r.StatusCode)
		})
	}
}
