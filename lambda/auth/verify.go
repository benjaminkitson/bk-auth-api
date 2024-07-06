package auth

import (
	"context"
	"encoding/json"

	"github.com/auth-api/lambda/secrets"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/jsii-runtime-go"
	"go.uber.org/zap"
)

type verifyInput struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type verifyOutput struct {
	Message string `json:"message"`
}

func HandleVerify(body string, sc secrets.SecretsClient, cc *cognitoidentityprovider.Client, logger *zap.Logger) (verifyOutput, error) {
	var v verifyInput
	json.Unmarshal([]byte(body), &v)

	clientId, err := sc.GetSecret("COGNITO_CLIENT")
	if err != nil {
		return verifyOutput{}, err
	}

	output, err := cc.ConfirmSignUp(context.TODO(), &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         jsii.String(clientId),
		ConfirmationCode: jsii.String(v.Code),
		Username:         jsii.String(v.Email),
	})

	if err != nil {
		logger.Error("failed!", zap.Error(err))
		return verifyOutput{}, err
	}

	logger.Info("verify output", zap.Any("output", output))
	return verifyOutput{
		Message: "Successfully verified email address!",
	}, nil
}
