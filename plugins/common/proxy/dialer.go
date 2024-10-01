package proxy

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/proxy"
)

type ProxiedDialer struct {
	dialer proxy.Dialer
}

func (pd *ProxiedDialer) Dial(network, addr string) (net.Conn, error) {
	return pd.dialer.Dial(network, addr)
}

func (pd *ProxiedDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if contextDialer, ok := pd.dialer.(proxy.ContextDialer); ok {
		return contextDialer.DialContext(ctx, network, addr)
	}

	contextDialer := contextDialerShim{pd.dialer}
	return contextDialer.DialContext(ctx, network, addr)
}

func (pd *ProxiedDialer) DialTimeout(network, addr string, timeout time.Duration) (net.Conn, error) {
	ctx := context.Background()
	if timeout.Seconds() != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	return pd.DialContext(ctx, network, addr)
}
