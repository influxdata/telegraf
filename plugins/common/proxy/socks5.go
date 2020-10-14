package proxy

import (
	"golang.org/x/net/proxy"
)

type Socks5ProxyConfig struct {
	address string    `toml:"socks5_address"`
	username string   `toml:"socks5_username"`
	password string   `toml:"socks5_password"`
}

func (c* Socks5ProxyConfig) GetDialer() (*proxy.Dialer, error) {
	var auth *proxy.Auth
	if c.username != "" || c.password != "" {
		auth = new(proxy.Auth)
		auth.User = c.username
		auth.Password = c.password
	}
	dialer, err := proxy.SOCKS5("tcp", c.address, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return &dialer, nil
}
