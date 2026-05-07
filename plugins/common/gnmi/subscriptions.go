package gnmi

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/gnxi/utils/xpath"
	"github.com/openconfig/gnmi/proto/gnmi"

	"github.com/influxdata/telegraf/config"
)

type Subscription struct {
	Name              string          `toml:"name"`
	Origin            string          `toml:"origin"`
	Path              string          `toml:"path"`
	SubscriptionMode  string          `toml:"subscription_mode"`
	SampleInterval    config.Duration `toml:"sample_interval"`
	SuppressRedundant bool            `toml:"suppress_redundant"`
	HeartbeatInterval config.Duration `toml:"heartbeat_interval"`

	fullPath *gnmi.Path
}

type TagSubscription struct {
	Match    string   `toml:"match"`
	Elements []string `toml:"elements"`
	Subscription
}

func (s *Subscription) Request() (*gnmi.Subscription, error) {
	gnmiPath, err := ParsePath(s.Origin, s.Path, "")
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

func (s *Subscription) Build(origin, prefix, target string) error {
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
