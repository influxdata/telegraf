// +build linux

package ethtool

import (
	"github.com/influxdata/telegraf/filter"
	"github.com/pkg/errors"
	"net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/safchain/ethtool"
)

type CommandEthtool struct {
	ethtool *ethtool.Ethtool
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

	for _, iface := range interfaces {

		// Check this isn't a loop back and that its matched by the filter
		if (iface.Flags&net.FlagLoopback == 0) && interfaceFilter.Match(iface.Name) {
			e.wg.Add(1)
			go e.gatherEthtoolStats(iface, acc)
		}
	}

	// Waiting for all the interfaces
	e.wg.Wait()
	return nil
}

// Gather the stats for the interface.
func (e *Ethtool) gatherEthtoolStats(iface net.Interface, acc telegraf.Accumulator) {
	defer e.wg.Done()

	tags := make(map[string]string)
	tags[tagInterface] = iface.Name

	// Optionally add driver name as a tag
	if e.DriverName {

		driverName, err := e.command.DriverName(iface.Name)
		if err != nil {
			driverErr := errors.Wrapf(err, "%s driver", iface.Name)
			acc.AddError(driverErr)
			return
		}

		tags[tagDriverName] = driverName
	}

	fields := make(map[string]interface{})
	stats, err := e.command.Stats(iface.Name)
	if err != nil {
		statsErr := errors.Wrapf(err, "%s stats", iface.Name)
		acc.AddError(statsErr)
		return
	}

	for k, v := range stats {
		fields[k] = v
	}

	acc.AddFields(pluginName, fields, tags)
}

func NewCommandEthtool() *CommandEthtool {
	e, _ := ethtool.NewEthtool()
	return &CommandEthtool{e}
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
			DriverName:       true,
			command:          NewCommandEthtool(),
		}
	})
}
