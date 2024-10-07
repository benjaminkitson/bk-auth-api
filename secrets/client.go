package secrets

import (
	"context"

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

/*
Simple wrapper around the Secrets Manager client for retrieving a secret
TODO: Write tests? Disproportionately complicated given how simple this is
*/
func NewSecretsClient(l *zap.Logger, sm *secretsmanager.Client) (SecretsClient, error) {
	return SecretsClient{
		logger: l,
		smc:    sm,
	}, nil

}

func (sc SecretsClient) GetSecret(name string) (string, error) {
	sc.logger.Info("Getting secret")
	sv, err := sc.smc.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: &name,
	})
	if err != nil {
		sc.logger.Error("Failed to retrieve secret", zap.Error(err))
		return "", err
	}
	return *sv.SecretString, nil
}
