package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/auth-api/lambda/secrets"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/jsii-runtime-go"
	"go.uber.org/zap"
)

type signUpInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpOutput struct {
	Message string `json:"message"`
}

func HandleSignUp(body string, sc secrets.SecretsClient, cc *cognitoidentityprovider.Client, logger *zap.Logger) (signUpOutput, error) {
	var s signUpInput
	json.Unmarshal([]byte(body), &s)

	clientId, err := sc.GetSecret("COGNITO_CLIENT")
	if err != nil {
		return signUpOutput{}, err
	}

	output, err := cc.SignUp(context.TODO(), &cognitoidentityprovider.SignUpInput{
		ClientId: jsii.String(clientId),
		Password: jsii.String(s.Password),
		Username: jsii.String(s.Email),

		// UserAttributes: []types.AttributeType{
		// 	{Name: jsii.String("email"), Value: jsii.String(s.Email)},
		// },
	})
	if err != nil {
		logger.Error("signup failed!", zap.Error(err))
		return signUpOutput{}, fmt.Errorf(err.Error())
	}
	logger.Info("signup output", zap.Any("output", output))

	return signUpOutput{
		Message: "Successfully signed up!",
	}, nil
}
