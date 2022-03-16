//go:build linux
// +build linux

package ethtool

import (
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
	ethtoolLib "github.com/safchain/ethtool"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type CommandEthtool struct {
	ethtool *ethtoolLib.Ethtool
}

func (e *Ethtool) Gather(acc telegraf.Accumulator) error {
	// Get the list of interfaces
	interfaces, err := e.command.Interfaces()
	if err != nil {
		acc.AddError(err)
		return nil
	}

	interfaceFilter, err := filter.NewIncludeExcludeFilter(e.InterfaceInclude, e.InterfaceExclude)
	if err != nil {
		return err
	}

	// parallelize the ethtool call in event of many interfaces
	var wg sync.WaitGroup

	for _, iface := range interfaces {
		// Check this isn't a loop back and that its matched by the filter
		if (iface.Flags&net.FlagLoopback == 0) && interfaceFilter.Match(iface.Name) {
			wg.Add(1)

			go func(i net.Interface) {
				e.gatherEthtoolStats(i, acc)
				wg.Done()
			}(iface)
		}
	}

	// Waiting for all the interfaces
	wg.Wait()
	return nil
}

// Initialise the Command Tool
func (e *Ethtool) Init() error {
	return e.command.Init()
}

// Gather the stats for the interface.
func (e *Ethtool) gatherEthtoolStats(iface net.Interface, acc telegraf.Accumulator) {
	tags := make(map[string]string)
	tags[tagInterface] = iface.Name

	driverName, err := e.command.DriverName(iface.Name)
	if err != nil {
		driverErr := errors.Wrapf(err, "%s driver", iface.Name)
		acc.AddError(driverErr)
		return
	}

	tags[tagDriverName] = driverName

	fields := make(map[string]interface{})
	stats, err := e.command.Stats(iface.Name)
	if err != nil {
		statsErr := errors.Wrapf(err, "%s stats", iface.Name)
		acc.AddError(statsErr)
		return
	}

	fields[fieldInterfaceUp] = e.interfaceUp(iface)
	for k, v := range stats {
		fields[e.normalizeKey(k)] = v
	}

	acc.AddFields(pluginName, fields, tags)
}

// normalize key string; order matters to avoid replacing whitespace with
// underscores, then trying to trim those same underscores. Likewise with
// camelcase before trying to lower case things.
func (e *Ethtool) normalizeKey(key string) string {
	// must trim whitespace or this will have a leading _
	if inStringSlice(e.NormalizeKeys, "snakecase") {
		key = camelCase2SnakeCase(strings.TrimSpace(key))
	}
	// must occur before underscore, otherwise nothing to trim
	if inStringSlice(e.NormalizeKeys, "trim") {
		key = strings.TrimSpace(key)
	}
	if inStringSlice(e.NormalizeKeys, "lower") {
		key = strings.ToLower(key)
	}
	if inStringSlice(e.NormalizeKeys, "underscore") {
		key = strings.ReplaceAll(key, " ", "_")
	}

	return key
}

func camelCase2SnakeCase(value string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(value, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func inStringSlice(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}

	return false
}

func (e *Ethtool) interfaceUp(iface net.Interface) bool {
	return (iface.Flags & net.FlagUp) != 0
}

func NewCommandEthtool() *CommandEthtool {
	return &CommandEthtool{}
}

func (c *CommandEthtool) Init() error {
	if c.ethtool != nil {
		return nil
	}

	e, err := ethtoolLib.NewEthtool()
	if err == nil {
		c.ethtool = e
	}

	return err
}

func (c *CommandEthtool) DriverName(intf string) (string, error) {
	return c.ethtool.DriverName(intf)
}

func (c *CommandEthtool) Stats(intf string) (map[string]uint64, error) {
	return c.ethtool.Stats(intf)
}

func (c *CommandEthtool) Interfaces() ([]net.Interface, error) {
	// Get the list of interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	return interfaces, nil
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		return &Ethtool{
			InterfaceInclude: []string{},
			InterfaceExclude: []string{},
			command:          NewCommandEthtool(),
		}
	})
}
