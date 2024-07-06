package auth

import "testing"

type mockSecretsClient struct {
	isSuccess  bool
	mockSecret string
}

func (msc mockSecretsClient) GetSecret(name string) (string, error) {
	return msc.mockSecret, nil
}

// type mockCognitoClient struct {

// }

func TestSignInSuccess(t *testing.T) {

}
