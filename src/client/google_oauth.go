package client

import (
	"errors"
	"go-fiber-template/domain/entities"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

// GoogleOAuthClient verifies Google ID tokens (credentials) issued to the
// frontend via Google's tokeninfo endpoint.
type GoogleOAuthClient struct {
	client   *resty.Client
	baseURL  string
	clientID string
}

type IGoogleOAuthClient interface {
	// VerifyIDToken validates a Google ID token and returns its claims.
	VerifyIDToken(idToken string) (*entities.GoogleTokenInfo, error)
}

func NewGoogleOAuthClient() IGoogleOAuthClient {
	baseURL := os.Getenv("GOOGLE_TOKENINFO_URL")
	if baseURL == "" {
		baseURL = "https://oauth2.googleapis.com/tokeninfo"
	}
	return &GoogleOAuthClient{
		client:   resty.New(),
		baseURL:  baseURL,
		clientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}
}

func (c *GoogleOAuthClient) VerifyIDToken(idToken string) (*entities.GoogleTokenInfo, error) {
	if idToken == "" {
		return nil, errors.New("id_token must not be empty")
	}

	info := &entities.GoogleTokenInfo{}
	resp, err := c.client.R().
		SetQueryParam("id_token", idToken).
		SetResult(info).
		Get(c.baseURL)

	if err != nil {
		log.Printf("GoogleOAuth -> VerifyIDToken: request error: %v\n", err)
		return nil, err
	}

	if resp.IsError() {
		log.Printf("GoogleOAuth -> VerifyIDToken: invalid token response: %s\n", resp.String())
		return nil, errors.New("invalid google id token")
	}

	if info.Sub == "" {
		return nil, errors.New("invalid google id token: missing subject")
	}

	// When GOOGLE_CLIENT_ID is configured, enforce the token audience so tokens
	// minted for other apps are rejected.
	if c.clientID != "" && info.Aud != c.clientID {
		return nil, errors.New("google id token audience mismatch")
	}

	return info, nil
}
