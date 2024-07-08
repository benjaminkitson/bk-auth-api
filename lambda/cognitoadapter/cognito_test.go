package cognito

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"go.uber.org/zap"
)

type MockCognitoClient struct {
	isError bool
}

var mockToken = "mockToken"

func (ma MockCognitoClient) InitiateAuth(ctx context.Context, params *cognitoidentityprovider.InitiateAuthInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	if ma.isError {
		return nil, fmt.Errorf("IntitiateAuth error")
	}
	return &cognitoidentityprovider.InitiateAuthOutput{
		AuthenticationResult: &types.AuthenticationResultType{
			AccessToken: &mockToken,
		},
	}, nil
}

func (ma MockCognitoClient) SignUp(ctx context.Context, params *cognitoidentityprovider.SignUpInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.SignUpOutput, error) {
	if ma.isError {
		return nil, fmt.Errorf("SignUp error")
	}
	return &cognitoidentityprovider.SignUpOutput{}, nil
}

func (ma MockCognitoClient) ConfirmSignUp(ctx context.Context, params *cognitoidentityprovider.ConfirmSignUpInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmSignUpOutput, error) {
	if ma.isError {
		return nil, fmt.Errorf("ConfirmSignUp error")
	}
	return &cognitoidentityprovider.ConfirmSignUpOutput{}, nil
}

/*
Tests the basic workings of the handler, using a mocked auth provider client that either succeeds or returns some generic error
*/
func TestSignIn(t *testing.T) {
	type test struct {
		Name             string
		RequestBody      map[string]string
		ExpectedError    bool
		ExpectedResponse map[string]string
	}

	tests := []test{
		{
			Name: "Sign in success",
			RequestBody: map[string]string{
				"email":    "abc@gmail.com",
				"password": "password",
			},
			ExpectedResponse: map[string]string{
				"message": signInSuccessMessage,
				"token":   mockToken,
			},
		},
		{
			Name: "Sign in error",
			RequestBody: map[string]string{
				"email":    "abc@gmail.com",
				"password": "password",
			},
			ExpectedError:    true,
			ExpectedResponse: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			l, err := zap.NewDevelopment()
			if err != nil {
				t.Fatalf("Failed to initialise dev logger")
			}

			m := MockCognitoClient{
				isError: tt.ExpectedError,
			}

			ca := NewAdapter(m, "MockClientId", l)
			if err != nil {
				t.Fatalf("Failed to initialise handler")
			}

			r, err := ca.SignIn(tt.RequestBody)
			if err != nil && !tt.ExpectedError {
				t.Fatalf("Unexpected handler error %v", err)
			}
			if !reflect.DeepEqual(r, tt.ExpectedResponse) {
				t.Fatalf("Unexpected response %v", r)
			}
		})
	}
}
