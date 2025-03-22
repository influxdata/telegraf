package kafka

import (
	"fmt"

	"github.com/IBM/sarama"

	"github.com/influxdata/telegraf/config"
)

type oauthToken struct {
	token      config.Secret
	extensions map[string]string
}

// Token does nothing smart, it just grabs a hard-coded token from config.
func (a *oauthToken) Token() (*sarama.AccessToken, error) {
	token, err := a.token.Get()
	if err != nil {
		return nil, fmt.Errorf("getting token failed: %w", err)
	}
	defer token.Destroy()
	return &sarama.AccessToken{
		Token:      token.String(),
		Extensions: a.extensions,
	}, nil
}
