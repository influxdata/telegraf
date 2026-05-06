package gnmi

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/gnxi/utils/xpath"
	"github.com/openconfig/gnmi/proto/gnmi"

	"github.com/influxdata/telegraf/config"
)

type subscription struct {
	Name              string          `toml:"name"`
	Origin            string          `toml:"origin"`
	Path              string          `toml:"path"`
	SubscriptionMode  string          `toml:"subscription_mode"`
	SampleInterval    config.Duration `toml:"sample_interval"`
	SuppressRedundant bool            `toml:"suppress_redundant"`
	HeartbeatInterval config.Duration `toml:"heartbeat_interval"`

	fullPath *gnmi.Path
}

type tagSubscription struct {
	subscription
	Match    string   `toml:"match"`
	Elements []string `toml:"elements"`
}

func (s *subscription) buildSubscription() (*gnmi.Subscription, error) {
	gnmiPath, err := parsePath(s.Origin, s.Path, "")
	if err != nil {
		return nil, err
	}
	mode, ok := gnmi.SubscriptionMode_value[strings.ToUpper(s.SubscriptionMode)]
	if !ok {
		return nil, fmt.Errorf("invalid subscription mode %s", s.SubscriptionMode)
	}
	return &gnmi.Subscription{
		Path:              gnmiPath,
		Mode:              gnmi.SubscriptionMode(mode),
		HeartbeatInterval: uint64(time.Duration(s.HeartbeatInterval).Nanoseconds()),
		SampleInterval:    uint64(time.Duration(s.SampleInterval).Nanoseconds()),
		SuppressRedundant: s.SuppressRedundant,
	}, nil
}

func (s *subscription) buildFullPath(origin, prefix, target string) error {
	var err error
	if s.fullPath, err = xpath.ToGNMIPath(s.Path); err != nil {
		return err
	}
	s.fullPath.Origin = s.Origin
	s.fullPath.Target = target
	if prefix != "" {
		prefixPath, err := xpath.ToGNMIPath(prefix)
		if err != nil {
			return err
		}
		s.fullPath.Elem = append(prefixPath.Elem, s.fullPath.Elem...)
		if s.Origin == "" && origin != "" {
			s.fullPath.Origin = origin
		}
	}
	return nil
}

func (s *subscription) buildAlias(enforceFirstNamespaceAsOrigin bool) (*pathInfo, string, error) {
	// Build the subscription path without keys
	path, err := parsePath(s.Origin, s.Path, "")
	if err != nil {
		return nil, "", err
	}
	info := newInfoFromPathWithoutKeys(path)
	if enforceFirstNamespaceAsOrigin {
		info.enforceFirstNamespaceAsOrigin()
	}

	// If the user didn't provide a measurement name, use last path element
	name := s.Name
	if name == "" && len(info.segments) > 0 {
		name = info.segments[len(info.segments)-1].id
	}
	return info, name, nil
}
