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

func (ma MockAdapter) SignIn(body string) (string, error) {
	if ma.isError {
		return "", fmt.Errorf("Auth provider error")
	}
	return "Sign in failed", nil
}

func (ma MockAdapter) SignUp(body string) (string, error) {
	if ma.isError {
		return "", fmt.Errorf("Auth provider error")
	}
	return "Sign up failed", nil
}

func (ma MockAdapter) VerifyEmail(body string) (string, error) {
	if ma.isError {
		return "", fmt.Errorf("Auth provider error")
	}
	return "Verify email failed", nil
}

type MockSecretsGetter struct {
	isError bool
}

func (ms MockSecretsGetter) GetSecret(name string) (string, error) {
	if ms.isError {
		return "", fmt.Errorf("Secrets client error")
	}
	return "superSecret", nil
}

func TestSignInSuccess(t *testing.T) {
	l, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to initialise dev logger")
	}

	s := MockSecretsGetter{
		isError: false,
	}

	m := MockAdapter{
		isError: false,
	}

	h, err := NewHandler(l, s, m)
	if err != nil {
		t.Fatalf("Failed to initialise handler")
	}

	req := events.APIGatewayProxyRequest{
		// This test should probably fail if the body isn't the correct format?
		Body: "{\"email\": \"abc@gmail.com\", \"password\": \"password\"}",
		Path: "/signin",
	}

	r, err := h.Handle(context.TODO(), req)
	if err != nil {
		t.Fatalf("Handler failed")
	}

	if r.StatusCode != 200 {
		t.Fatalf("Expected Status Code to be 200")
	}
}
