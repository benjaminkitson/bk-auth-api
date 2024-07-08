package cognito

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/jsii-runtime-go"
	"go.uber.org/zap"
)

type Adapter struct {
	identityProviderClient *cognitoidentityprovider.Client
	clientId               string
	logger                 *zap.Logger
}

func NewAdapter(cc *cognitoidentityprovider.Client, ccid string, logger *zap.Logger) Adapter {
	return Adapter{
		identityProviderClient: cc,
		clientId:               ccid,
		logger:                 logger,
	}
}

// TODO: All of the "input" and "output" types in this file could probably just be `map[string]string`
// TODO: Some errors (username already exists, incorrect password etc) aren't really errors at all, and need to be accounted for

/*
Sign in
*/

type signInOutput struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

func (ca Adapter) SignIn(body map[string]string) (string, error) {
	output, err := ca.identityProviderClient.InitiateAuth(context.TODO(), &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       "USER_PASSWORD_AUTH",
		ClientId:       aws.String(ca.clientId),
		AuthParameters: map[string]string{"USERNAME": body["email"], "PASSWORD": body["password"]},
	})

	if err != nil {
		ca.logger.Error("signup failed!", zap.Error(err))
		return "", fmt.Errorf(err.Error())
	}
	ca.logger.Info("signin output", zap.Any("output", output))

	// TODO: this seems contrived
	r := signInOutput{
		Token:   *output.AuthenticationResult.AccessToken,
		Message: "Successfully signed in!",
	}

	return ca.getResponseBody(r), nil
}

/*
Sign up
*/

type signUpOutput struct {
	Message string `json:"message"`
}

func (ca Adapter) SignUp(body map[string]string) (string, error) {
	output, err := ca.identityProviderClient.SignUp(context.TODO(), &cognitoidentityprovider.SignUpInput{
		ClientId: jsii.String(ca.clientId),
		Password: jsii.String(body["password"]),
		Username: jsii.String(body["email"]),

		// UserAttributes: []types.AttributeType{
		// 	{Name: jsii.String("email"), Value: jsii.String(s.Email)},
		// },
	})
	if err != nil {
		ca.logger.Error("signup failed!", zap.Error(err))
		return "", fmt.Errorf(err.Error())
	}
	ca.logger.Info("signup output", zap.Any("output", output))

	// TODO: this seems contrived
	r := signUpOutput{
		Message: "Successfully signed up!",
	}

	return ca.getResponseBody(r), nil
}

type verifyOutput struct {
	Message string `json:"message"`
}

/*
Verify email
*/

func (ca Adapter) VerifyEmail(body map[string]string) (string, error) {
	output, err := ca.identityProviderClient.ConfirmSignUp(context.TODO(), &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         jsii.String(ca.clientId),
		ConfirmationCode: jsii.String(body["code"]),
		Username:         jsii.String(body["email"]),
	})

	if err != nil {
		ca.logger.Error("failed!", zap.Error(err))
		return "", err
	}

	ca.logger.Info("verify output", zap.Any("output", output))

	// TODO: this seems contrived
	r := verifyOutput{
		Message: "Successfully verified email address!",
	}

	return ca.getResponseBody(r), nil
}

// Wrapper around json.Marshal that handles generic error cases
func (ca Adapter) getResponseBody(data any) string {
	// Fallback JSON string in case marshalling somehow goes wrong
	e := "{\"message\": \"Something went wrong!\"}"
	if _, ok := data.(error); ok {
		// If the input to getResponseBody is an error, return the generic error message
		return e
	}
	r, err := json.Marshal(data)
	if err != nil {
		ca.logger.Error("Failed to get response body from Cognito adapter", zap.Error(err))
		return e
	}
	return string(r)
}
