package auth

import (
	"context"
	"errors"
	"fmt"

	vault "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/aws"
)

type AwsIAM struct {
	RoleName       string `toml:"role_name"`
	Region         string `toml:"region"`
	ServerIDHeader string `toml:"server_id_header"`
}

// Init validates the auth method options and sets any necessary defaults
func (a *AwsIAM) Init() error {
	if a.RoleName == "" {
		return errors.New("aws iam role_name missing")
	}

	if a.Region == "" {
		a.Region = "us-east-1"
	}

	return nil
}

// Authenticate uses the provided configuration to authenticate to Vault
func (a *AwsIAM) Authenticate(v *vault.Client) (*vault.Secret, error) {
	opts := []aws.LoginOption{
		aws.WithIAMAuth(),
		aws.WithRole(a.RoleName),
		aws.WithRegion(a.Region),
	}
	if a.ServerIDHeader != "" {
		opts = append(opts, aws.WithIAMServerIDHeader(a.ServerIDHeader))
	}

	awsAuth, err := aws.NewAWSAuth(opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS IAM auth method: %w", err)
	}

	authInfo, err := v.Auth().Login(context.Background(), awsAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AWS IAM auth method: %w", err)
	}
	if authInfo == nil {
		return nil, errors.New("no auth info was returned after login")
	}

	return authInfo, nil
}

type AwsEC2 struct {
	RoleName      string `toml:"role_name"`
	Region        string `toml:"region"`
	SignatureType string `toml:"signature_type"`
}

// Init validates the auth method options and sets any necessary defaults
func (a *AwsEC2) Init() error {
	if a.RoleName == "" {
		return errors.New("aws ec2 role_name missing")
	}

	switch a.SignatureType {
	case "":
		a.SignatureType = "pkcs7"
	case "pkcs7", "identity", "rsa2048":
	default:
		return fmt.Errorf("unknown signature type: %q", a.SignatureType)
	}

	if a.Region == "" {
		a.Region = "us-east-1"
	}

	return nil
}

// Authenticate uses the provided configuration to authenticate to Vault
func (a *AwsEC2) Authenticate(v *vault.Client) (*vault.Secret, error) {
	opts := []aws.LoginOption{
		aws.WithEC2Auth(),
		aws.WithRole(a.RoleName),
		aws.WithRegion(a.Region),
	}

	switch a.SignatureType {
	case "pkcs7":
		opts = append(opts, aws.WithPKCS7Signature())
	case "identity":
		opts = append(opts, aws.WithIdentitySignature())
	case "rsa2048":
		opts = append(opts, aws.WithRSA2048Signature())
	}

	awsAuth, err := aws.NewAWSAuth(opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize AWS EC2 auth method: %w", err)
	}

	authInfo, err := v.Auth().Login(context.Background(), awsAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to login to AWS EC2 auth method: %w", err)
	}
	if authInfo == nil {
		return nil, errors.New("no auth info was returned after login")
	}

	return authInfo, nil
}
