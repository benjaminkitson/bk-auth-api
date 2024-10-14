package cognito

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/jsii-runtime-go"
	"go.uber.org/zap"
)

type Adapter struct {
	// identityProviderClient cognitoidentityprovider.Client
	identityProviderClient CognitoClient
	clientId               string
	logger                 *zap.Logger
	userPoolID             string
}

func NewAdapter(cc CognitoClient, ccid string, u string, logger *zap.Logger) Adapter {
	return Adapter{
		identityProviderClient: cc,
		clientId:               ccid,
		logger:                 logger,
		userPoolID:             u,
	}
}

type CognitoClient interface {
	InitiateAuth(context.Context, *cognitoidentityprovider.InitiateAuthInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error)
	SignUp(context.Context, *cognitoidentityprovider.SignUpInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.SignUpOutput, error)
	ConfirmSignUp(context.Context, *cognitoidentityprovider.ConfirmSignUpInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmSignUpOutput, error)
	AdminDeleteUser(context.Context, *cognitoidentityprovider.AdminDeleteUserInput, ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.AdminDeleteUserOutput, error)
}

// TODO: Some errors (username already exists, incorrect password etc) aren't really errors at all, and need to be accounted for
// TODO: All of these methods should accept ctx as the first arg

const signInSuccessMessage = "Successfully signed in!"

func (ca Adapter) SignIn(body map[string]string) (map[string]string, error) {
	if body["email"] == "" || body["password"] == "" {
		ca.logger.Error("invalid request body!")
		return nil, fmt.Errorf("invalid request body")
	}

	output, err := ca.identityProviderClient.InitiateAuth(context.Background(), &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       "USER_PASSWORD_AUTH",
		ClientId:       aws.String(ca.clientId),
		AuthParameters: map[string]string{"USERNAME": body["email"], "PASSWORD": body["password"]},
	})

	if err != nil {
		ca.logger.Error("signup failed!", zap.Error(err))
		return nil, fmt.Errorf(err.Error())
	}
	ca.logger.Info("signin output", zap.Any("output", output))

	return map[string]string{
		"token":   *output.AuthenticationResult.AccessToken,
		"message": signInSuccessMessage,
	}, nil
}

const signUpSuccessMessage = "Successfully signed up!"

func (ca Adapter) SignUp(body map[string]string) (map[string]string, error) {
	if body["email"] == "" || body["password"] == "" {
		ca.logger.Error("invalid request body!")
		return nil, fmt.Errorf("invalid request body")
	}

	output, err := ca.identityProviderClient.SignUp(context.Background(), &cognitoidentityprovider.SignUpInput{
		ClientId: jsii.String(ca.clientId),
		Password: jsii.String(body["password"]),
		Username: jsii.String(body["email"]),
	})

	if err != nil {
		ca.logger.Error("signup failed!", zap.Error(err))
		return nil, fmt.Errorf(err.Error())
	}
	ca.logger.Info("signup output", zap.Any("output", output))

	return map[string]string{
		"message": signUpSuccessMessage,
	}, nil
}

const verifyEmailSuccessMessage = "Successfully verified email address!"

func (ca Adapter) VerifyEmail(body map[string]string) (map[string]string, error) {
	if body["email"] == "" || body["code"] == "" {
		ca.logger.Error("invalid request body!")
		return nil, fmt.Errorf("invalid request body")
	}

	output, err := ca.identityProviderClient.ConfirmSignUp(context.Background(), &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         jsii.String(ca.clientId),
		ConfirmationCode: jsii.String(body["code"]),
		Username:         jsii.String(body["email"]),
	})

	if err != nil {
		ca.logger.Error("failed!", zap.Error(err))
		return nil, err
	}

	ca.logger.Info("verify output", zap.Any("output", output))

	return map[string]string{
		"message": verifyEmailSuccessMessage,
	}, nil
}

const adminDeleteSuccessMessage = "Successfully deleted user from auth provider"

// TODO: The map[string]string breaks down a bit for this one - consider making it a special case?
func (ca Adapter) AdminDelete(body map[string]string) (map[string]string, error) {
	if body["email"] == "" {
		ca.logger.Error("invalid request body!")
		return nil, fmt.Errorf("invalid request body")
	}

	email := body["email"]
	// TODO: Investigate what exactly is in the metadata, and if it's needed

	// TODO: Change this to the user pool id
	_, err := ca.identityProviderClient.AdminDeleteUser(context.Background(), &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: &ca.userPoolID,
		Username:   &email,
	})
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"message": adminDeleteSuccessMessage,
	}, nil
}
