package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/auth-api/lambda/secrets"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"go.uber.org/zap"
)

type signInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signInOutput struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

func HandleSignIn(body string, sc secrets.SecretsClient, cc *cognitoidentityprovider.Client, logger *zap.Logger) (signInOutput, error) {
	var s signInInput
	// TODO: errors aren't handled in any of these cases
	json.Unmarshal([]byte(body), &s)

	clientId, err := sc.GetSecret("COGNITO_CLIENT")
	if err != nil {
		return signInOutput{}, err
	}

	output, err := cc.InitiateAuth(context.TODO(), &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       "USER_PASSWORD_AUTH",
		ClientId:       aws.String(clientId),
		AuthParameters: map[string]string{"USERNAME": s.Email, "PASSWORD": s.Password},
	})

	if err != nil {
		logger.Error("signup failed!", zap.Error(err))
		return signInOutput{}, fmt.Errorf(err.Error())
	}
	logger.Info("signin output", zap.Any("output", output))

	return signInOutput{
		Token:   *output.AuthenticationResult.AccessToken,
		Message: "Successfully signed in!",
	}, nil
}
