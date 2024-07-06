package secrets

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"
)

type SecretsClient struct {
	logger *zap.Logger
	smc    *secretsmanager.Client
}

type SecretGetter interface {
	GetSecret(string) (string, error)
}

func NewSecretsClient(l *zap.Logger) (SecretsClient, error) {
	l.Info("Initialising secrets client")
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		l.Error("Failed to intialise SDK config", zap.Error(err))
		return SecretsClient{}, err
	}
	s := secretsmanager.NewFromConfig(sdkConfig)
	return SecretsClient{
		logger: l,
		smc:    s,
	}, nil

}

func (sc SecretsClient) GetSecret(name string) (string, error) {
	sv, err := sc.smc.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &name,
	})
	if err != nil {
		return "", err
	}
	return *sv.SecretString, nil
}
