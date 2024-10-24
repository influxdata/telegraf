//go:build linux

package ethtool

import (
	"math"
	"net"
	"runtime"

	"github.com/safchain/ethtool"
	"github.com/vishvananda/netns"

	"github.com/influxdata/telegraf"
)

type namespace interface {
	name() string
	interfaces() ([]namespacedInterface, error)
	driverName(intf namespacedInterface) (string, error)
	stats(intf namespacedInterface) (map[string]uint64, error)
	get(intf namespacedInterface) (map[string]uint64, error)
}

type namespacedInterface struct {
	net.Interface
	namespace namespace
}

type namespacedAction struct {
	result chan<- namespacedResult
	f      func(*namespaceGoroutine) (interface{}, error)
}

type namespacedResult struct {
	result interface{}
	err    error
}

type namespaceGoroutine struct {
	namespaceName string
	handle        netns.NsHandle
	ethtoolClient *ethtool.Ethtool
	c             chan namespacedAction
	log           telegraf.Logger
}

func (n *namespaceGoroutine) name() string {
	return n.namespaceName
}

func (n *namespaceGoroutine) interfaces() ([]namespacedInterface, error) {
	interfaces, err := n.do(func(n *namespaceGoroutine) (interface{}, error) {
		interfaces, err := net.Interfaces()
		if err != nil {
			return nil, err
		}
		namespacedInterfaces := make([]namespacedInterface, 0, len(interfaces))
		for _, iface := range interfaces {
			namespacedInterfaces = append(
				namespacedInterfaces,
				namespacedInterface{
					Interface: iface,
					namespace: n,
				},
			)
		}
		return namespacedInterfaces, nil
	})

	return interfaces.([]namespacedInterface), err
}

func (n *namespaceGoroutine) driverName(intf namespacedInterface) (string, error) {
	driver, err := n.do(func(n *namespaceGoroutine) (interface{}, error) {
		return n.ethtoolClient.DriverName(intf.Name)
	})
	return driver.(string), err
}

func (n *namespaceGoroutine) stats(intf namespacedInterface) (map[string]uint64, error) {
	driver, err := n.do(func(n *namespaceGoroutine) (interface{}, error) {
		return n.ethtoolClient.Stats(intf.Name)
	})
	return driver.(map[string]uint64), err
}

func (n *namespaceGoroutine) get(intf namespacedInterface) (map[string]uint64, error) {
	result, err := n.do(func(n *namespaceGoroutine) (interface{}, error) {
		ecmd := ethtool.EthtoolCmd{}
		speed32, err := n.ethtoolClient.CmdGet(&ecmd, intf.Name)
		if err != nil {
			return nil, err
		}

		var speed = uint64(speed32)
		if speed == math.MaxUint32 {
			speed = math.MaxUint64
		}

		var link32 uint32
		link32, err = n.ethtoolClient.LinkState(intf.Name)
		if err != nil {
			return nil, err
		}

		return map[string]uint64{
			"speed":   speed,
			"duplex":  uint64(ecmd.Duplex),
			"autoneg": uint64(ecmd.Autoneg),
			"link":    uint64(link32),
		}, nil
	})

	if result != nil {
		return result.(map[string]uint64), err
	}
	return nil, err
}

// start locks a goroutine to an OS thread and ties it to the namespace, then
// loops for actions to run in the namespace.
func (n *namespaceGoroutine) start() error {
	n.c = make(chan namespacedAction)
	started := make(chan error)
	go func() {
		// We're going to hold this thread locked permanently. We're going to
		// do this for every namespace. This makes it very likely that the Go
		// runtime will spin up new threads to replace it. To avoid thread
		// leaks, we don't unlock when we're done and instead let this thread
		// die.
		runtime.LockOSThread()

		// If this goroutine is for the initial namespace, we are already in
		// the correct namespace. Switching would require CAP_SYS_ADMIN, which
		// we may not have. Don't switch if the desired namespace matches the
		// current one.
		initialNamespace, err := netns.Get()
		if err != nil {
			n.log.Errorf("Could not get initial namespace: %s", err)
			started <- err
			return
		}
		if !initialNamespace.Equal(n.handle) {
			if err := netns.Set(n.handle); err != nil {
				n.log.Errorf("Could not switch to namespace %q: %s", n.namespaceName, err.Error())
				started <- err
				return
			}
		}

		// Every namespace needs its own connection to ethtool
		e, err := ethtool.NewEthtool()
		if err != nil {
			n.log.Errorf("Could not create ethtool client for namespace %q: %s", n.namespaceName, err.Error())
			started <- err
			return
		}
		n.ethtoolClient = e
		started <- nil
		for command := range n.c {
			result, err := command.f(n)
			command.result <- namespacedResult{
				result: result,
				err:    err,
			}
			close(command.result)
		}
	}()
	return <-started
}

// do runs a function inside the OS thread tied to the namespace.
func (n *namespaceGoroutine) do(f func(*namespaceGoroutine) (interface{}, error)) (interface{}, error) {
	result := make(chan namespacedResult)
	n.c <- namespacedAction{
		result: result,
		f:      f,
	}
	r := <-result
	return r.result, r.err
}
