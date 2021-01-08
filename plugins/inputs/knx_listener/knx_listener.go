package knx_listener

import (
	"fmt"
	"reflect"

	"github.com/vapourismo/knx-go/knx"
	"github.com/vapourismo/knx-go/knx/dpt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type KNXInterface interface {
	Inbound() <-chan knx.GroupEvent
	Close()
}

type addressTarget struct {
	measurement string
	datapoint   dpt.DatapointValue
}

type Measurement struct {
	Name      string
	Dpt       string
	Addresses []string
}

type KNXListener struct {
	ServiceType    string          `toml:"service_type"`
	ServiceAddress string          `toml:"service_address"`
	Measurements   []Measurement   `toml:"measurement"`
	Log            telegraf.Logger `toml:"-"`

	client      KNXInterface
	gaTargetMap map[string]addressTarget

	acc telegraf.Accumulator
}

func (kl *KNXListener) Description() string {
	return "Listener capable of handling KNX bus messages provided through a KNX-IP Interface."
}

func (kl *KNXListener) SampleConfig() string {
	return `
  ## Type of KNX-IP interface.
  ## Can be either "tunnel" or "router".
  # service_type = "tunnel"

  ## Address of the KNX-IP interface.
  service_address = "localhost:3671"

  ## Measurement definition(s)
  # [[inputs.KNXListener.measurement]]
  #   ## Name of the measurement
  #   name = "temperature"
  #   ## Datapoint-Type (DPT) of the KNX messages
  #   dpt = "9.001"
  #   ## List of Group-Addresses (GAs) assigned to the measurement
  #   addresses = ["5/5/1"]

  # [[inputs.KNXListener.measurement]]
  #   name = "illumination"
  #   dpt = "9.004"
  #   addresses = ["5/5/3"]
`
}

func (kl *KNXListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (kl *KNXListener) Start(acc telegraf.Accumulator) error {
	// Store the accumulator for later use
	kl.acc = acc

	// Construct the mapping of Group-addresses (GAs) to DPTs and the name
	// of the measurement
	kl.gaTargetMap = make(map[string]addressTarget)
	for _, m := range kl.Measurements {
		kl.Log.Debugf("Group-address mapping for measurement %q:", m.Name)
		for _, ga := range m.Addresses {
			kl.Log.Debugf("  %v --> %s", ga, m.Dpt)
			if _, ok := kl.gaTargetMap[ga]; ok {
				return fmt.Errorf("duplicate specification of address %v", ga)
			}
			d, ok := dpt.Produce(m.Dpt)
			if !ok {
				return fmt.Errorf("cannot create datapoint-type %v for address %v", m.Dpt, ga)
			}
			kl.gaTargetMap[ga] = addressTarget{m.Name, d}
		}
	}

	// Connect to the KNX-IP interface
	kl.Log.Infof("Trying to connect to %q at %q", kl.ServiceType, kl.ServiceAddress)
	switch kl.ServiceType {
	case "tunnel":
		c, err := knx.NewGroupTunnel(kl.ServiceAddress, knx.DefaultTunnelConfig)
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
		c, err := NewDummyInterface()
		if err != nil {
			return err
		}
		kl.client = &c
	default:
		return fmt.Errorf("invalid interface type: %s", kl.ServiceAddress)
	}
	kl.Log.Infof("Connected!")

	// Listen to the KNX bus
	go kl.listen()

	return nil
}

func (kl *KNXListener) Stop() {
	if kl.client != nil {
		kl.client.Close()
	}
}

func (kl *KNXListener) listen() {
	for msg := range kl.client.Inbound() {
		// Match GA to DataPointType and measurement name
		ga := msg.Destination.String()
		target, ok := kl.gaTargetMap[ga]
		if ok {
			err := target.datapoint.Unpack(msg.Data)
			if err != nil {
				kl.Log.Errorf("Unpacking data failed: %v", err)
				continue
			}
			kl.Log.Debugf("Matched GA %q to measurement %q with value %v", ga, target.measurement, target.datapoint)

			// Convert the DatapointValue interface back to its basic type again
			// as otherwise telegraf will not push out the metrics and eat it
			// silently.
			var value interface{}
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
			default:
				kl.Log.Errorf("Type conversion %v failed for address %v", ga, vi.Kind())
				continue
			}

			// Compose the actual data to be pushed out
			fields := map[string]interface{}{"value": value}
			tags := map[string]string{
				"groupaddress": ga,
				"unit":         target.datapoint.(dpt.DatapointMeta).Unit(),
				"source":       msg.Source.String(),
			}
			kl.acc.AddFields(target.measurement, fields, tags)
		} else {
			kl.Log.Infof("Ignoring message %+v for unknown GA %q", msg, ga)
		}
	}
}

func init() {
	inputs.Add("KNXListener", func() telegraf.Input { return &KNXListener{ServiceType: "tunnel"} })
}
