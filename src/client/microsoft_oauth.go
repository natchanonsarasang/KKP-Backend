package client

import (
	"errors"
	"go-fiber-template/domain/entities"
	"log"
	"os"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type MicrosoftOAuthClient struct {
	jwksURL  string
	clientID string
	jwks     *keyfunc.JWKS
}

type IMicrosoftOAuthClient interface {
	VerifyIDToken(idToken string) (*entities.MicrosoftTokenInfo, error)
}

func NewMicrosoftOAuthClient() IMicrosoftOAuthClient {
	jwksURL := os.Getenv("MICROSOFT_JWKS_URL")
	if jwksURL == "" {
		jwksURL = "https://login.microsoftonline.com/common/discovery/v2.0/keys"
	}
	clientID := os.Getenv("MICROSOFT_CLIENT_ID")

	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		log.Printf("MicrosoftOAuth -> NewMicrosoftOAuthClient: failed to fetch JWKS from %s: %v\n", jwksURL, err)
	}

	return &MicrosoftOAuthClient{
		jwksURL:  jwksURL,
		clientID: clientID,
		jwks:     jwks,
	}
}

func (c *MicrosoftOAuthClient) VerifyIDToken(idToken string) (*entities.MicrosoftTokenInfo, error) {
	if idToken == "" {
		return nil, errors.New("id_token must not be empty")
	}

	if c.jwks == nil {
		var err error
		c.jwks, err = keyfunc.Get(c.jwksURL, keyfunc.Options{})
		if err != nil {
			log.Printf("MicrosoftOAuth -> VerifyIDToken: retry fetching JWKS failed: %v\n", err)
			return nil, errors.New("unable to retrieve Microsoft public keys for verification")
		}
	}

	token, err := jwt.Parse(idToken, c.jwks.Keyfunc)
	if err != nil {
		log.Printf("MicrosoftOAuth -> VerifyIDToken: JWT parse error: %v\n", err)
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid microsoft id token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid microsoft id token claims")
	}

	info := &entities.MicrosoftTokenInfo{}

	if aud, ok := claims["aud"].(string); ok {
		info.Aud = aud
	}
	if sub, ok := claims["sub"].(string); ok {
		info.Sub = sub
	}
	if email, ok := claims["email"].(string); ok {
		info.Email = email
	} else if prefUsername, ok := claims["preferred_username"].(string); ok {
		info.Email = prefUsername
	}
	if name, ok := claims["name"].(string); ok {
		info.Name = name
	}
	if prefUsername, ok := claims["preferred_username"].(string); ok {
		info.PreferredUsername = prefUsername
	}
	if oid, ok := claims["oid"].(string); ok {
		info.Oid = oid
	}

	if info.Sub == "" {
		return nil, errors.New("invalid microsoft id token: missing subject")
	}

	if c.clientID != "" && info.Aud != c.clientID {
		return nil, errors.New("microsoft id token audience mismatch")
	}

	return info, nil
}
