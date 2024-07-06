package cognito

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/auth-api/lambda/secrets"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/jsii-runtime-go"
	"go.uber.org/zap"
)

type CognitoAdapter struct {
	identityProviderClient *cognitoidentityprovider.Client
	clientId               string
	logger                 *zap.Logger
}

func NewCognitoAdapter(sc secrets.SecretGetter, logger *zap.Logger) (CognitoAdapter, error) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Error("Failed to intialise SDK config", zap.Error(err))
		return CognitoAdapter{}, nil
	}

	cc := cognitoidentityprovider.NewFromConfig(sdkConfig)
	ccid, err := sc.GetSecret("COGNITO_CLIENT")
	if err != nil {
		logger.Error("Faled to get cognito client id", zap.Error(err))
		return CognitoAdapter{}, nil
	}

	return CognitoAdapter{
		identityProviderClient: cc,
		clientId:               ccid,
		logger:                 logger,
	}, nil
}

type CognitoAdapterHandler func(string) (string, error)

// TODO: All of the "input" and "output" types in this file could probably just be `map[string]string``

/*
Sign in
*/

type signInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInOutput struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

func (ca CognitoAdapter) SignIn(body string) (string, error) {
	var s signInInput
	// TODO: errors aren't handled in any of these cases
	err := json.Unmarshal([]byte(body), &s)
	if err != nil {
		return "", err
	}

	output, err := ca.identityProviderClient.InitiateAuth(context.TODO(), &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       "USER_PASSWORD_AUTH",
		ClientId:       aws.String(ca.clientId),
		AuthParameters: map[string]string{"USERNAME": s.Email, "PASSWORD": s.Password},
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

type signUpInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpOutput struct {
	Message string `json:"message"`
}

func (ca CognitoAdapter) SignUp(body string) (string, error) {
	var s signUpInput
	err := json.Unmarshal([]byte(body), &s)
	if err != nil {
		return "", err
	}

	output, err := ca.identityProviderClient.SignUp(context.TODO(), &cognitoidentityprovider.SignUpInput{
		ClientId: jsii.String(ca.clientId),
		Password: jsii.String(s.Password),
		Username: jsii.String(s.Email),

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

type verifyInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type verifyOutput struct {
	Message string `json:"message"`
}

/*
Verify email
*/

func (ca CognitoAdapter) VerifyEmail(body string) (string, error) {
	var v verifyInput
	err := json.Unmarshal([]byte(body), &v)
	if err != nil {
		return "", err
	}

	output, err := ca.identityProviderClient.ConfirmSignUp(context.TODO(), &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         jsii.String(ca.clientId),
		ConfirmationCode: jsii.String(v.Code),
		Username:         jsii.String(v.Email),
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
func (ca CognitoAdapter) getResponseBody(data any) string {
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
