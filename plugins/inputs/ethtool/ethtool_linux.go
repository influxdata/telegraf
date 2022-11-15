//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package ethtool

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/vishvananda/netns"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type CommandEthtool struct {
	Log                 telegraf.Logger
	namespaceGoroutines map[string]*NamespaceGoroutine
}

func (e *Ethtool) Init() error {
	var err error
	e.interfaceFilter, err = filter.NewIncludeExcludeFilter(e.InterfaceInclude, e.InterfaceExclude)
	if err != nil {
		return err
	}

	if e.DownInterfaces == "" {
		e.DownInterfaces = "expose"
	}

	if err = choice.Check(e.DownInterfaces, downInterfacesBehaviors); err != nil {
		return fmt.Errorf("down_interfaces: %w", err)
	}

	// If no namespace include or exclude filters were provided, then default
	// to just the initial namespace.
	e.includeNamespaces = len(e.NamespaceInclude) > 0 || len(e.NamespaceExclude) > 0
	if len(e.NamespaceInclude) == 0 && len(e.NamespaceExclude) == 0 {
		e.NamespaceInclude = []string{""}
	} else if len(e.NamespaceInclude) == 0 {
		e.NamespaceInclude = []string{"*"}
	}
	e.namespaceFilter, err = filter.NewIncludeExcludeFilter(e.NamespaceInclude, e.NamespaceExclude)
	if err != nil {
		return err
	}

	if command, ok := e.command.(*CommandEthtool); ok {
		command.Log = e.Log
	}

	return e.command.Init()
}

func (e *Ethtool) Gather(acc telegraf.Accumulator) error {
	// Get the list of interfaces
	interfaces, err := e.command.Interfaces(e.includeNamespaces)
	if err != nil {
		acc.AddError(err)
		return nil
	}

	// parallelize the ethtool call in event of many interfaces
	var wg sync.WaitGroup

	for _, iface := range interfaces {
		// Check this isn't a loop back and that its matched by the filter(s)
		if e.interfaceEligibleForGather(iface) {
			wg.Add(1)

			go func(i NamespacedInterface) {
				e.gatherEthtoolStats(i, acc)
				wg.Done()
			}(iface)
		}
	}

	// Waiting for all the interfaces
	wg.Wait()
	return nil
}

func (e *Ethtool) interfaceEligibleForGather(iface NamespacedInterface) bool {
	// Don't gather if it is a loop back, or it isn't matched by the filter
	if isLoopback(iface) || !e.interfaceFilter.Match(iface.Name) {
		return false
	}

	// Don't gather if it's not in a namespace matched by the filter
	if !e.namespaceFilter.Match(iface.Namespace.Name()) {
		return false
	}

	// For downed interfaces, gather only for "expose"
	if !interfaceUp(iface) {
		return e.DownInterfaces == "expose"
	}

	return true
}

// Gather the stats for the interface.
func (e *Ethtool) gatherEthtoolStats(iface NamespacedInterface, acc telegraf.Accumulator) {
	tags := make(map[string]string)
	tags[tagInterface] = iface.Name
	tags[tagNamespace] = iface.Namespace.Name()

	driverName, err := e.command.DriverName(iface)
	if err != nil {
		driverErr := errors.Wrapf(err, "%s driver", iface.Name)
		acc.AddError(driverErr)
		return
	}

	tags[tagDriverName] = driverName

	fields := make(map[string]interface{})
	stats, err := e.command.Stats(iface)
	if err != nil {
		statsErr := errors.Wrapf(err, "%s stats", iface.Name)
		acc.AddError(statsErr)
		return
	}

	fields[fieldInterfaceUp] = interfaceUp(iface)
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
	// aws has a conflicting name that needs to be renamed
	if key == "interface_up" {
		key = "interface_up_counter"
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

func isLoopback(iface NamespacedInterface) bool {
	return (iface.Flags & net.FlagLoopback) != 0
}

func interfaceUp(iface NamespacedInterface) bool {
	return (iface.Flags & net.FlagUp) != 0
}

func NewCommandEthtool() *CommandEthtool {
	return &CommandEthtool{}
}

func (c *CommandEthtool) Init() error {
	// Create the goroutine for the initial namespace
	initialNamespace, err := netns.Get()
	if err != nil {
		return err
	}
	namespaceGoroutine := &NamespaceGoroutine{
		name:   "",
		handle: initialNamespace,
		Log:    c.Log,
	}
	if err := namespaceGoroutine.Start(); err != nil {
		c.Log.Errorf(`Failed to start goroutine for the initial namespace: %s`, err)
		return err
	}
	c.namespaceGoroutines = map[string]*NamespaceGoroutine{
		"": namespaceGoroutine,
	}
	return nil
}

func (c *CommandEthtool) DriverName(intf NamespacedInterface) (driver string, err error) {
	return intf.Namespace.DriverName(intf)
}

func (c *CommandEthtool) Stats(intf NamespacedInterface) (stats map[string]uint64, err error) {
	return intf.Namespace.Stats(intf)
}

func (c *CommandEthtool) Interfaces(includeNamespaces bool) ([]NamespacedInterface, error) {
	const namespaceDirectory = "/var/run/netns"

	initialNamespace, err := netns.Get()
	if err != nil {
		c.Log.Errorf("Could not get initial namespace: %s", err)
		return nil, err
	}

	// Gather the list of namespace names to from which to retrieve interfaces.
	initialNamespaceIsNamed := false
	var namespaceNames []string
	// Handles are only used to create namespaced goroutines. We don't prefill
	// with the handle for the initial namespace because we've already created
	// its goroutine in Init().
	handles := map[string]netns.NsHandle{}

	if includeNamespaces {
		namespaces, err := os.ReadDir(namespaceDirectory)
		if err != nil {
			c.Log.Warnf("Could not find namespace directory: %s", err)
		}

		// We'll always have at least the initial namespace, so add one to ensure
		// we have capacity for it.
		namespaceNames = make([]string, 0, len(namespaces)+1)
		for _, namespace := range namespaces {
			name := namespace.Name()
			namespaceNames = append(namespaceNames, name)

			handle, err := netns.GetFromPath(filepath.Join(namespaceDirectory, name))
			if err != nil {
				c.Log.Warnf(`Could not get handle for namespace "%s": %s`, name, err)
				continue
			}
			handles[name] = handle
			if handle.Equal(initialNamespace) {
				initialNamespaceIsNamed = true
			}
		}
	}

	// We don't want to gather interfaces from the same namespace twice, and
	// it's possible, though unlikely, that the initial namespace is also a
	// named interface.
	if !initialNamespaceIsNamed {
		namespaceNames = append(namespaceNames, "")
	}

	allInterfaces := make([]NamespacedInterface, 0)
	for _, namespace := range namespaceNames {
		if _, ok := c.namespaceGoroutines[namespace]; !ok {
			c.namespaceGoroutines[namespace] = &NamespaceGoroutine{
				name:   namespace,
				handle: handles[namespace],
				Log:    c.Log,
			}
			if err := c.namespaceGoroutines[namespace].Start(); err != nil {
				c.Log.Errorf(`Failed to start goroutine for namespace "%s": %s`, namespace, err)
				delete(c.namespaceGoroutines, namespace)
				continue
			}
		}

		interfaces, err := c.namespaceGoroutines[namespace].Interfaces()
		if err != nil {
			c.Log.Warnf(`Could not get interfaces from namespace "%s": %s`, namespace, err)
			continue
		}
		allInterfaces = append(allInterfaces, interfaces...)
	}

	return allInterfaces, nil
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		return &Ethtool{
			InterfaceInclude: []string{},
			InterfaceExclude: []string{},
			NamespaceInclude: []string{},
			NamespaceExclude: []string{},
			command:          NewCommandEthtool(),
		}
	})
}
