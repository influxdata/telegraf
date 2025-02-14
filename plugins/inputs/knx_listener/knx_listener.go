//go:generate ../../../tools/readme_config_includer/generator
package knx_listener

import (
	_ "embed"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type KNXListener struct {
	ServiceType    string          `toml:"service_type"`
	ServiceAddress string          `toml:"service_address"`
	Measurements   []measurement   `toml:"measurement"`
	Log            telegraf.Logger `toml:"-"`

	client      knxInterface
	gaTargetMap map[string]addressTarget
	gaLogbook   map[string]bool

	wg        sync.WaitGroup
	connected atomic.Bool
}

type measurement struct {
	Name      string   `toml:"name"`
	Dpt       string   `toml:"dpt"`
	AsString  bool     `toml:"as_string"`
	Addresses []string `toml:"addresses"`
}

type addressTarget struct {
	measurement string
	asstring    bool
	datapoint   dpt.Datapoint
}

type knxInterface interface {
	Inbound() <-chan knx.GroupEvent
	Close()
}

func (*KNXListener) SampleConfig() string {
	return sampleConfig
}

func (kl *KNXListener) Init() error {
	// Setup a logbook to track unknown GAs to avoid log-spamming
	kl.gaLogbook = make(map[string]bool)

	// Construct the mapping of Group-addresses (GAs) to DPTs and the name
	// of the measurement
	kl.gaTargetMap = make(map[string]addressTarget)
	for _, m := range kl.Measurements {
		kl.Log.Debugf("Group-address mapping for measurement %q:", m.Name)
		for _, ga := range m.Addresses {
			kl.Log.Debugf("  %s --> %s", ga, m.Dpt)
			if _, ok := kl.gaTargetMap[ga]; ok {
				return fmt.Errorf("duplicate specification of address %q", ga)
			}
			d, ok := dpt.Produce(m.Dpt)
			if !ok {
				return fmt.Errorf("cannot create datapoint-type %q for address %q", m.Dpt, ga)
			}
			kl.gaTargetMap[ga] = addressTarget{measurement: m.Name, asstring: m.AsString, datapoint: d}
		}
	}

	return nil
}

func (kl *KNXListener) Start(acc telegraf.Accumulator) error {
	// Connect to the KNX-IP interface
	kl.Log.Infof("Trying to connect to %q at %q", kl.ServiceType, kl.ServiceAddress)
	switch kl.ServiceType {
	case "tunnel", "tunnel_udp":
		tunnelconfig := knx.DefaultTunnelConfig
		tunnelconfig.UseTCP = false
		c, err := knx.NewGroupTunnel(kl.ServiceAddress, tunnelconfig)
		if err != nil {
			return err
		}
		kl.client = &c
	case "tunnel_tcp":
		tunnelconfig := knx.DefaultTunnelConfig
		tunnelconfig.UseTCP = true
		c, err := knx.NewGroupTunnel(kl.ServiceAddress, tunnelconfig)
		if err != nil {
			return err
		}
		kl.client = &c
	case "router":
		c, err := knx.NewGroupRouter(kl.ServiceAddress, knx.DefaultRouterConfig)
		if err != nil {
			return err
		}
		kl.client = &c
	case "dummy":
		c := newDummyInterface()
		kl.client = &c
	default:
		return fmt.Errorf("invalid interface type: %s", kl.ServiceAddress)
	}
	kl.Log.Infof("Connected!")
	kl.connected.Store(true)

	// Listen to the KNX bus
	kl.wg.Add(1)
	go func() {
		defer kl.wg.Done()
		kl.listen(acc)
		kl.connected.Store(false)
		acc.AddError(errors.New("disconnected from bus"))
	}()

	return nil
}

func (kl *KNXListener) Gather(acc telegraf.Accumulator) error {
	if !kl.connected.Load() {
		// We got disconnected for some reason, so try to reconnect in every
		// gather cycle until we are reconnected
		if err := kl.Start(acc); err != nil {
			return fmt.Errorf("reconnecting to bus failed: %w", err)
		}
	}

	return nil
}

func (kl *KNXListener) Stop() {
	if kl.client != nil {
		kl.client.Close()
		kl.wg.Wait()
	}
}

func (kl *KNXListener) listen(acc telegraf.Accumulator) {
	for msg := range kl.client.Inbound() {
		if msg.Command == knx.GroupRead {
			// Ignore GroupValue_Read requests as they would either
			// - fail to unpack due to invalid data length (DPT != 1) or
			// - create invalid `false` values as their data always unpacks `0` (DPT 1)
			continue
		}
		// Match GA to DataPointType and measurement name
		ga := msg.Destination.String()
		target, ok := kl.gaTargetMap[ga]
		if !ok {
			if !kl.gaLogbook[ga] {
				kl.Log.Infof("Ignoring message %+v for unknown GA %q", msg, ga)
				kl.gaLogbook[ga] = true
			}
			continue
		}

		// Extract the value from the data-frame
		if err := target.datapoint.Unpack(msg.Data); err != nil {
			kl.Log.Errorf("Unpacking data failed: %v", err)
			continue
		}
		kl.Log.Debugf("Matched GA %q to measurement %q with value %v", ga, target.measurement, target.datapoint)

		// Convert the DatapointValue interface back to its basic type again
		// as otherwise telegraf will not push out the metrics and eat it
		// silently.
		var value interface{}
		if !target.asstring {
			vi := reflect.Indirect(reflect.ValueOf(target.datapoint))
			switch vi.Kind() {
			case reflect.Bool:
				value = vi.Bool()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value = vi.Int()
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value = vi.Uint()
			case reflect.Float32, reflect.Float64:
				value = vi.Float()
			case reflect.String:
				value = vi.String()
			default:
				kl.Log.Errorf("Type conversion %v failed for address %q", vi.Kind(), ga)
				continue
			}
		} else {
			value = target.datapoint.String()
		}

		// Compose the actual data to be pushed out
		fields := map[string]interface{}{"value": value}
		tags := map[string]string{
			"groupaddress": ga,
			"unit":         target.datapoint.(dpt.DatapointMeta).Unit(),
			"source":       msg.Source.String(),
		}
		acc.AddFields(target.measurement, fields, tags)
	}
}

func init() {
	inputs.Add("knx_listener", func() telegraf.Input { return &KNXListener{ServiceType: "tunnel"} })
	// Register for backward compatibility
	inputs.Add("KNXListener", func() telegraf.Input { return &KNXListener{ServiceType: "tunnel"} })
}
