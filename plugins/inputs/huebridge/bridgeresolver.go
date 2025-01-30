package huebridge

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/tdrn-org/go-hue"
)

type bridgeResolver struct {
	bridgeUrl          *url.URL
	cachedBridgeClient hue.BridgeClient
}

func newBridgeResolver(rawUrl string) (*bridgeResolver, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bridge URL %q (reason: %w)", rawUrl, err)
	}
	return &bridgeResolver{bridgeUrl: parsedUrl}, nil
}

func (resolver *bridgeResolver) String() string {
	return resolver.bridgeUrl.Redacted()
}

func (resolver *bridgeResolver) reset() {
	resolver.cachedBridgeClient = nil
}

func (resolver *bridgeResolver) resolveBridge(rcc *RemoteClientConfig, tcc *tls.ClientConfig, timeout config.Duration) (hue.BridgeClient, error) {
	if resolver.cachedBridgeClient != nil {
		return resolver.cachedBridgeClient, nil
	}
	var err error
	switch resolver.bridgeUrl.Scheme {
	case "address":
		resolver.cachedBridgeClient, err = resolver.resolveBridgeViaAddress(timeout)
	case "cloud":
		resolver.cachedBridgeClient, err = resolver.resolveBridgeViaCloud(tcc, timeout)
	case "mdns":
		resolver.cachedBridgeClient, err = resolver.resolveBridgeViaMDNS(timeout)
	case "remote":
		resolver.cachedBridgeClient, err = resolver.resolveBridgeViaRemote(rcc, tcc, timeout)
	default:
		return nil, fmt.Errorf("unrecognized bridge URL %q", resolver)
	}
	return resolver.cachedBridgeClient, err
}

func (resolver *bridgeResolver) resolveBridgeViaAddress(timeout config.Duration) (hue.BridgeClient, error) {
	locator, err := hue.NewAddressBridgeLocator(resolver.bridgeUrl.Host)
	if err != nil {
		return nil, err
	}
	return resolver.resolveLocalBridge(locator, timeout)
}

func (resolver *bridgeResolver) resolveBridgeViaCloud(tcc *tls.ClientConfig, timeout config.Duration) (hue.BridgeClient, error) {
	locator := hue.NewCloudBridgeLocator()
	if resolver.bridgeUrl.Host != "" {
		discoveryEndpointUrl, err := url.Parse(fmt.Sprintf("https://%s/", resolver.bridgeUrl.Host))
		if err != nil {
			return nil, err
		}
		discoveryEndpointUrl = discoveryEndpointUrl.JoinPath(resolver.bridgeUrl.Path)
		locator.DiscoveryEndpointUrl = discoveryEndpointUrl
	}
	tlsConfig, err := tcc.TLSConfig()
	if err != nil {
		return nil, err
	}
	locator.TlsConfig = tlsConfig
	return resolver.resolveLocalBridge(locator, timeout)
}

func (resolver *bridgeResolver) resolveBridgeViaMDNS(timeout config.Duration) (hue.BridgeClient, error) {
	locator := hue.NewMDNSBridgeLocator()
	return resolver.resolveLocalBridge(locator, timeout)
}

func (resolver *bridgeResolver) resolveLocalBridge(locator hue.BridgeLocator, timeout config.Duration) (hue.BridgeClient, error) {
	bridge, err := locator.Lookup(resolver.bridgeUrl.User.Username(), time.Duration(timeout))
	if err != nil {
		return nil, err
	}
	bridgeUrlPassword, set := resolver.bridgeUrl.User.Password()
	if !set {
		return nil, fmt.Errorf("no password set in bridge URL %q", resolver.bridgeUrl)
	}
	return bridge.NewClient(hue.NewLocalBridgeAuthenticator(bridgeUrlPassword), time.Duration(timeout))
}

func (resolver *bridgeResolver) resolveBridgeViaRemote(rcc *RemoteClientConfig, tcc *tls.ClientConfig, timeout config.Duration) (hue.BridgeClient, error) {
	if rcc.RemoteClientId == "" || rcc.RemoteClientSecret == "" || rcc.RemoteTokenDir == "" {
		return nil, fmt.Errorf("remote application credentials and/or token director not configured")
	}
	var redirectUrl *url.URL
	if rcc.RemoteCallbackUrl != "" {
		parsedRedirectUrl, err := url.Parse(rcc.RemoteCallbackUrl)
		if err != nil {
			return nil, err
		}
		redirectUrl = parsedRedirectUrl
	}
	tokenFile := filepath.Join(rcc.RemoteTokenDir, rcc.RemoteClientId, strings.ToUpper(resolver.bridgeUrl.User.Username())+".json")
	locator, err := hue.NewRemoteBridgeLocator(rcc.RemoteClientId, rcc.RemoteClientSecret, redirectUrl, tokenFile)
	if err != nil {
		return nil, err
	}
	if resolver.bridgeUrl.Host != "" {
		endpointUrl, err := url.Parse(fmt.Sprintf("https://%s/", resolver.bridgeUrl.Host))
		if err != nil {
			return nil, err
		}
		endpointUrl = endpointUrl.JoinPath(resolver.bridgeUrl.Path)
		locator.EndpointUrl = endpointUrl
	}
	tlsConfig, err := tcc.TLSConfig()
	if err != nil {
		return nil, nil
	}
	locator.TlsConfig = tlsConfig
	return resolver.resolveRemoteBridge(locator, timeout)
}

func (resolver *bridgeResolver) resolveRemoteBridge(locator *hue.RemoteBridgeLocator, timeout config.Duration) (hue.BridgeClient, error) {
	bridge, err := locator.Lookup(resolver.bridgeUrl.User.Username(), time.Duration(timeout))
	if err != nil {
		return nil, err
	}
	bridgeUrlPassword, set := resolver.bridgeUrl.User.Password()
	if !set {
		return nil, fmt.Errorf("no password set in bridge URL %q", resolver.bridgeUrl)
	}
	return bridge.NewClient(hue.NewRemoteBridgeAuthenticator(locator, bridgeUrlPassword), time.Duration(timeout))
}
