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

// TODO: Some errors (username already exists, incorrect password etc) aren't really errors at all, and need to be accounted for

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
		"message": "Successfully signed in!",
	}, nil
}

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
		"message": "Successfully signed up!",
	}, nil
}

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
		"message": "Successfully verified email address!",
	}, nil
}
