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
	identityProviderClient CognitoClient
	clientId               string
	logger                 *zap.Logger
}

func NewAdapter(cc CognitoClient, ccid string, logger *zap.Logger) Adapter {
	return Adapter{
		identityProviderClient: cc,
		clientId:               ccid,
		logger:                 logger,
	}
}

type CognitoClient interface {
	InitiateAuth(ctx context.Context, params *cognitoidentityprovider.InitiateAuthInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.InitiateAuthOutput, error)
	SignUp(ctx context.Context, params *cognitoidentityprovider.SignUpInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.SignUpOutput, error)
	ConfirmSignUp(ctx context.Context, params *cognitoidentityprovider.ConfirmSignUpInput, optFns ...func(*cognitoidentityprovider.Options)) (*cognitoidentityprovider.ConfirmSignUpOutput, error)
}

// TODO: Some errors (username already exists, incorrect password etc) aren't really errors at all, and need to be accounted for

const signInSuccessMessage = "Successfully signed in!"

func (ca Adapter) SignIn(body map[string]string) (map[string]string, error) {
	output, err := ca.identityProviderClient.InitiateAuth(context.TODO(), &cognitoidentityprovider.InitiateAuthInput{
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
		return nil, fmt.Errorf(err.Error())
	}
	ca.logger.Info("signup output", zap.Any("output", output))

	return map[string]string{
		"message": signUpSuccessMessage,
	}, nil
}

const verifyEmailSuccessMessage = "Successfully verified email address!"

func (ca Adapter) VerifyEmail(body map[string]string) (map[string]string, error) {
	output, err := ca.identityProviderClient.ConfirmSignUp(context.TODO(), &cognitoidentityprovider.ConfirmSignUpInput{
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
