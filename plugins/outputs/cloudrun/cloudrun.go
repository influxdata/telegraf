package cloudrun

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	jwtGo "github.com/golang-jwt/jwt/v4"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"

	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

const sampleConfig = `
  ## A plugin that can transmit metrics over OAuth2
  ## URL is the Cloud Run Wavefront proxy address to send metrics to
  # url = "http://127.0.0.1:8080/telegraf"

  ## Timeout for Cloud Run message, suggested as 30s to account for handshaking
  # timeout = "30s"

  ## Cloud Run JSON file location
  ## This is the location of the JSON file generated from your GCP project that's authorized to send
  ## metrics into CloudRun.
  ## Windows users, note that you need to use forward slashes.
  # credentials_file = "/etc/telegraf/example_secret.json"

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "wavefront"

  ## NOTE: The default headers have already been set to the following by default:
  ## defaultContentType   = "application/octet-stream"
  ## defaultAccept        = "application/json"
  ## defaultMethod        = http.MethodPost
`

const (
	defaultClientTimeout = 5 * time.Second
	defaultContentType   = "application/octet-stream"
	defaultAccept        = "application/json"
	defaultMethod        = http.MethodPost
)

type CloudRun struct {
	URL                 string          `toml:"url"`
	CredentialsFile     string          `toml:"credentials_file"`
	DisableConvertPaths bool            `toml:"wavefront_disable_path_conversion"`
	Log                 telegraf.Logger `toml:"-"`
	httpconfig.HTTPClientConfig

	client      *http.Client
	serializer  serializers.Serializer
	accessToken string
}

func (cr *CloudRun) SetSerializer(serializer serializers.Serializer) {
	cr.serializer = serializer
}

func (cr *CloudRun) Connect() error {
	if cr.client == nil {
		return cr.setUpDefaultClient()
	}

	return nil
}

func (cr *CloudRun) setUpDefaultClient() error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cr.Timeout))
	defer cancel()

	err := cr.getAccessToken(ctx)
	if err != nil {
		return err
	}

	client, err := cr.HTTPClientConfig.CreateClient(ctx, cr.Log)
	if err != nil {
		return err
	}

	cr.client = client
	return nil
}

func (cr *CloudRun) getAccessToken(ctx context.Context) error {
	data, err := ioutil.ReadFile(cr.CredentialsFile)
	if err != nil {
		return err
	}

	conf, err := google.JWTConfigFromJSON(data, cr.URL)
	if err != nil {
		return err
	}

	jwtConfig := &jwt.Config{
		Email:         conf.Email,
		TokenURL:      conf.TokenURL,
		PrivateKey:    conf.PrivateKey,
		PrivateClaims: map[string]interface{}{"target_audience": cr.URL},
	}
	token, err := jwtConfig.TokenSource(ctx).Token()
	if err != nil {
		return err
	}

	cr.accessToken = token.Extra("id_token").(string)
	return nil
}

func (cr *CloudRun) Close() error {
	return nil
}

func (cr *CloudRun) Description() string {
	return "A plugin that is capable of transmitting metrics over HTTPS to a metrics proxy hosted in Cloud Run"
}

func (cr *CloudRun) SampleConfig() string {
	return sampleConfig
}

func (cr *CloudRun) Write(metrics []telegraf.Metric) error {
	reqBody, err := cr.serializer.SerializeBatch(metrics)
	if err != nil {
		return err
	}

	return cr.send(reqBody)
}

func (cr *CloudRun) send(reqBody []byte) error {
	var reqBodyBuffer io.Reader = bytes.NewBuffer(reqBody)
	var err error
	req, err := http.NewRequest(defaultMethod, cr.URL, reqBodyBuffer)
	if err != nil {
		return err
	}

	// Inspect jwt claims to view expiration time
	claims := jwtGo.RegisteredClaims{}
	jwtGo.ParseWithClaims(cr.accessToken, &claims, func(token *jwtGo.Token) (interface{}, error) {
		return nil, nil
	})

	// Request new token if expired
	if !claims.VerifyExpiresAt(time.Now(), true) {
		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Duration(cr.Timeout))
		defer cancel()

		err = cr.getAccessToken(ctx)
		if err != nil {
			return err
		}
	}

	bearerToken := "Bearer " + cr.accessToken

	req.Header.Set("User-Agent", internal.ProductToken())
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", bearerToken)

	resp, err := cr.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if _, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("when writing to %q received status code	: %d", cr.URL, resp.StatusCode)
	}

	return nil
}

func init() {
	outputs.Add("cloudrun", func() telegraf.Output {
		return &CloudRun{}
	})
}
