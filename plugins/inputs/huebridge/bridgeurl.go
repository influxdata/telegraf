package huebridge

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/tdrn-org/go-hue"
)

type BridgeURL url.URL

func ParseBridgeURL(rawUrl string) (*BridgeURL, error) {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bridge URL %q (reason: %w)", rawUrl, err)
	}
	return (*BridgeURL)(parsedUrl), nil
}

func (bridgeUrl *BridgeURL) String() string {
	return (*url.URL)(bridgeUrl).Redacted()
}

func (bridgeUrl *BridgeURL) ResolveBridge(bcc *BridgeClientConfig) (hue.BridgeClient, error) {
	switch bridgeUrl.Scheme {
	case "address":
		return bridgeUrl.ResolveBridgeViaAddress(bcc)
	case "cloud":
		return bridgeUrl.ResolveBridgeViaCloud(bcc)
	case "mdns":
		return bridgeUrl.ResolveBridgeViaMDNS(bcc)
	case "remote":
		return bridgeUrl.ResolveBridgeViaRemote(bcc)
	}
	return nil, fmt.Errorf("unrecognized bridge URL %q", bridgeUrl)
}

func (bridgeUrl *BridgeURL) ResolveBridgeViaAddress(bcc *BridgeClientConfig) (hue.BridgeClient, error) {
	locator, err := hue.NewAddressBridgeLocator(bridgeUrl.Host)
	if err != nil {
		return nil, err
	}
	return bridgeUrl.ResolveLocalBridge(locator, bcc)
}

func (bridgeUrl *BridgeURL) ResolveBridgeViaCloud(bcc *BridgeClientConfig) (hue.BridgeClient, error) {
	locator := hue.NewCloudBridgeLocator()
	if bridgeUrl.Host != "" {
		discoveryEndpointUrl, err := url.Parse(fmt.Sprintf("https://%s/", bridgeUrl.Host))
		if err != nil {
			return nil, err
		}
		discoveryEndpointUrl = discoveryEndpointUrl.JoinPath(bridgeUrl.Path)
		locator.DiscoveryEndpointUrl = discoveryEndpointUrl
	}
	tlsConfig, err := bcc.TLSConfig()
	if err != nil {
		return nil, nil
	}
	locator.TlsConfig = tlsConfig
	return bridgeUrl.ResolveLocalBridge(locator, bcc)
}

func (bridgeUrl *BridgeURL) ResolveBridgeViaMDNS(bcc *BridgeClientConfig) (hue.BridgeClient, error) {
	locator := hue.NewMDNSBridgeLocator()
	locator.Limit = 1
	return bridgeUrl.ResolveLocalBridge(locator, bcc)
}

func (bridgeUrl *BridgeURL) ResolveLocalBridge(locator hue.BridgeLocator, bcc *BridgeClientConfig) (hue.BridgeClient, error) {
	bridge, err := locator.Lookup(bridgeUrl.User.Username(), time.Duration(bcc.Timeout))
	if err != nil {
		return nil, err
	}
	bridgeUrlPassword, set := bridgeUrl.User.Password()
	if !set {
		return nil, fmt.Errorf("no password set in bridge URL %q", bridgeUrl)
	}
	return bridge.NewClient(hue.NewLocalBridgeAuthenticator(bridgeUrlPassword), time.Duration(bcc.Timeout))
}

func (bridgeUrl *BridgeURL) ResolveBridgeViaRemote(bcc *BridgeClientConfig) (hue.BridgeClient, error) {
	if bcc.RemoteClientId == "" || bcc.RemoteClientSecret == "" || bcc.RemoteTokenDir == "" {
		return nil, fmt.Errorf("remote application credentials and/or token director not configured")
	}
	var redirectUrl *url.URL
	if bcc.RemoteCallbackUrl != "" {
		parsedRedirectUrl, err := url.Parse(bcc.RemoteCallbackUrl)
		if err != nil {
			return nil, err
		}
		redirectUrl = parsedRedirectUrl
	}
	tokenFile := filepath.Join(bcc.RemoteTokenDir, bcc.RemoteClientId, strings.ToUpper(bridgeUrl.User.Username())+".json")
	locator, err := hue.NewRemoteBridgeLocator(bcc.RemoteClientId, bcc.RemoteClientSecret, redirectUrl, tokenFile)
	if err != nil {
		return nil, err
	}
	if bridgeUrl.Host != "" {
		endpointUrl, err := url.Parse(fmt.Sprintf("https://%s/", bridgeUrl.Host))
		if err != nil {
			return nil, err
		}
		endpointUrl = endpointUrl.JoinPath(bridgeUrl.Path)
		locator.EndpointUrl = endpointUrl
	}
	tlsConfig, err := bcc.TLSConfig()
	if err != nil {
		return nil, nil
	}
	locator.TlsConfig = tlsConfig
	return bridgeUrl.ResolveRemoteBridge(locator, bcc)
}

func (bridgeUrl *BridgeURL) ResolveRemoteBridge(locator *hue.RemoteBridgeLocator, bcc *BridgeClientConfig) (hue.BridgeClient, error) {
	bridge, err := locator.Lookup(bridgeUrl.User.Username(), time.Duration(bcc.Timeout))
	if err != nil {
		return nil, err
	}
	bridgeUrlPassword, set := bridgeUrl.User.Password()
	if !set {
		return nil, fmt.Errorf("no password set in bridge URL %q", bridgeUrl)
	}
	return bridge.NewClient(hue.NewRemoteBridgeAuthenticator(locator, bridgeUrlPassword), time.Duration(bcc.Timeout))
}
