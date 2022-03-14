package proto

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/pkg/errors"
)

type CognitoIdentityProvider struct {
	awsCip   *cognitoidentityprovider.CognitoIdentityProvider
	user     *string
	password *string
	clientId *string

	accessToken  *string
	idToken      *string
	refreshToken *string
	expiresAt    time.Time
}

func NewCognitoIdentityProvider(sess *session.Session, user, password, clientID *string) *CognitoIdentityProvider {
	awsCip := cognitoidentityprovider.New(sess)
	return &CognitoIdentityProvider{
		awsCip:    awsCip,
		user:      user,
		password:  password,
		clientId:  clientID,
		expiresAt: time.Now().Add(-1 * time.Minute),
	}
}

func (c *CognitoIdentityProvider) GetAccessToken() (*string, error) {
	if c.accessToken != nil && !c.isExpired() {
		return c.accessToken, nil
	}

	if err := c.refreshTokenLogin(); err != nil {
		if err := c.userPasswordLogin(); err != nil {
			return nil, err
		}
	}

	return c.accessToken, nil
}

func (c *CognitoIdentityProvider) GetIdAccessToken() (*string, error) {
	if c.idToken != nil && !c.isExpired() {
		return c.idToken, nil
	}

	if err := c.refreshTokenLogin(); err != nil {
		if err := c.userPasswordLogin(); err != nil {
			return nil, err
		}
	}

	return c.idToken, nil
}

func (c *CognitoIdentityProvider) userPasswordLogin() error {
	resp, err := c.awsCip.InitiateAuth(
		c.loginRequest(cognitoidentityprovider.AuthFlowTypeUserPasswordAuth, map[string]*string{
			"USERNAME": c.user,
			"PASSWORD": c.password,
		}))
	if err != nil {
		return errors.Wrap(err, "unable to get access token using user/password auth")
	}
	c.accessToken = resp.AuthenticationResult.AccessToken
	c.idToken = resp.AuthenticationResult.IdToken
	c.refreshToken = resp.AuthenticationResult.RefreshToken
	c.expiresAt = time.Now().Add(time.Duration(*resp.AuthenticationResult.ExpiresIn) * time.Second)

	return nil
}

func (c *CognitoIdentityProvider) refreshTokenLogin() error {
	if c.refreshToken == nil {
		return fmt.Errorf("refresh token is empty")
	}

	if c.isExpired() {
		return fmt.Errorf("refresh token is expired")
	}

	resp, err := c.awsCip.InitiateAuth(
		c.loginRequest(cognitoidentityprovider.AuthFlowTypeRefreshTokenAuth, map[string]*string{
			"REFRESH_TOKEN": c.refreshToken,
		}))
	if err != nil {
		return errors.Wrap(err, "unable to get access token using refresh token")
	}
	c.accessToken = resp.AuthenticationResult.AccessToken
	c.idToken = resp.AuthenticationResult.IdToken
	c.refreshToken = resp.AuthenticationResult.RefreshToken
	c.expiresAt = time.Now().Add(time.Duration(*resp.AuthenticationResult.ExpiresIn) * time.Second)

	return nil
}

func (c *CognitoIdentityProvider) loginRequest(authFlow string, authParams map[string]*string) *cognitoidentityprovider.InitiateAuthInput {
	return &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       aws.String(authFlow),
		AuthParameters: authParams,
		ClientId:       c.clientId,
	}
}

func (c *CognitoIdentityProvider) isExpired() bool {
	return c.expiresAt.Add(-1 * time.Minute).After(time.Now())
}
