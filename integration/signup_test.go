package signup_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"testing"

	"github.com/antihax/optional"
	"github.com/benjaminkitson/bk-auth-api/integration/env"
	mail "github.com/mailslurp/mailslurp-client-go"
	"github.com/stretchr/testify/assert"
)

func AuthPost(p string, body map[string]string) error {
	r, err := url.Parse(env.AuthURL)
	if err != nil {
		return err
	}
	r.Path = path.Join(r.Path, p)
	b, err := json.Marshal(body)
	br := bytes.NewReader(b)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, r.String(), br)
	if err != nil {
		return err
	}

	c := http.Client{}

	res, err := c.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode == 200 {
		return nil
	}

	return fmt.Errorf("unexpected status code %v", res.StatusCode)
}

func GetVerificationCode() (string, error) {
	ctx := context.WithValue(
		context.Background(),
		mail.ContextAPIKey,
		mail.APIKey{Key: env.MailAPIKey},
	)

	config := mail.NewConfiguration()
	client := mail.NewAPIClient(config)

	waitOpts := &mail.WaitForLatestEmailOpts{
		InboxId:    optional.NewInterface(env.MailAPIKey),
		Timeout:    optional.NewInt64(30000),
		UnreadOnly: optional.NewBool(true),
	}
	email, _, err := client.WaitForControllerApi.WaitForLatestEmail(ctx, waitOpts)
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(env.MailString)
	code := r.FindStringSubmatch(*email.Body)[1]

	return code, nil
}

func TestSignUp(t *testing.T) {
	type test struct {
		Name               string
		AdapterError       bool
		RequestBody        string
		ExpectedStatusCode int
	}

	tests := []test{
		{
			Name:               "Sign in success",
			ExpectedStatusCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := AuthPost("signup", map[string]string{
				"email":    env.TestEmail,
				"password": "Password123",
			})

			if err != nil {
				t.Log(err.Error())
			}
			assert.NoError(t, err)

			c, err := GetVerificationCode()
			if err != nil {
				t.Log(err.Error())
			}
			assert.NoError(t, err)

			err = AuthPost("verify", map[string]string{
				"email": env.TestEmail,
				"code":  c,
			})
			if err != nil {
				t.Log(err.Error())
			}
			assert.NoError(t, err)
		})
	}
}
