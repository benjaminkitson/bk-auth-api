package integration

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
	"github.com/benjaminkitson/bk-user-api/models"
	mail "github.com/mailslurp/mailslurp-client-go"
	"github.com/stretchr/testify/assert"
)

func AuthPost(p string, body map[string]string) (models.User, error) {
	r, err := url.Parse(env.AuthURL)
	if err != nil {
		return models.User{}, err
	}
	r.Path = path.Join(r.Path, p)
	b, err := json.Marshal(body)
	br := bytes.NewReader(b)
	if err != nil {
		return models.User{}, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, r.String(), br)
	if err != nil {
		return models.User{}, err
	}

	c := http.Client{}

	res, err := c.Do(req)
	if err != nil {
		return models.User{}, err
	}

	if res.StatusCode == 200 {
		var u models.User
		err = json.NewDecoder(res.Body).Decode(&u)
		if err != nil {
			return models.User{}, err
		}
		return u, nil
	}

	return models.User{}, fmt.Errorf("unexpected status code %v", res.StatusCode)
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
		InboxId:    optional.NewInterface(env.InboxID),
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
			t.Log("Attempting sign up...")
			_, err := AuthPost("signup", map[string]string{
				"email":    env.TestEmail,
				"password": "Password123",
			})

			if err != nil {
				t.Log(err.Error())
			}
			assert.NoError(t, err)
			t.Log("Sign up succesfull!")

			t.Log("Getting email verification code...")
			c, err := GetVerificationCode()
			if err != nil {
				t.Log(err.Error())
			}
			assert.NoError(t, err)
			t.Logf("Retrieved code %v", c)

			t.Log("Verifying email address...")
			u, err := AuthPost("verify", map[string]string{
				"email": env.TestEmail,
				"code":  c,
			})
			if err != nil {
				t.Log(err.Error())
			}
			assert.NoError(t, err)
			t.Logf("Verified email address for user %v", u.UserID)

			t.Logf("Cleaning up created user...")
			r, err := AuthPost("admin-delete", map[string]string{
				"id":    u.UserID,
				"email": u.Email,
			})
			if err != nil {
				t.Log(err.Error())
			}
			assert.NoError(t, err)
			t.Logf("Deleted data for user %v", r.UserID)
		})
	}
}
