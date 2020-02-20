package knx_listener

import (
	"log"

	"github.com/vapourismo/knx-go/knx/dpt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

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
	ServiceType    string        `toml:"service_type"`
	ServiceAddress string        `toml:"service_address"`
	Measurements   []Measurement `toml:"measurement"`

	client      KNXInterface
	gaTargetMap map[string]addressTarget

	acc telegraf.Accumulator
}

func (kl *KNXListener) Description() string {
	return "Listener capable of handling KNX bus messages provided through a KNX-IP Interface."
}

func (kl *KNXListener) SampleConfig() string {
	return `
  # Type of KNX-IP interface.
  # Can be either "tunnel", "router" or "dummy" (for testing only).
  service_type = "tunnel"

  # Address of the KNX-IP interface.
  service_address = "localhost:3671"

  ## Measurement definition(s)
  # [[inputs.KNXListener.measurement]]
  #   # Name of the measurement
  #   name = "temperature"
  #   # Datapoint-Type (DPT) of the KNX messages
  #   dpt = "9.001"
  #   # List of Group-Addresses (GAs) assigned to the measurement
  #   addresses = ["5/5/1"]

  # [[inputs.KNXListener.measurement]]
  #   name = "illumination"
  #   dpt = "9.004"
  #   addresses = ["5/5/3"]
`
}

func (kl *KNXListener) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (kl *KNXListener) Start(acc telegraf.Accumulator) error {
	// Store the accumulator for later use
	kl.acc = acc

	// Construct the mapping of Group-addresses (GAs) to DPTs and the name
	// of the measurement
	kl.gaTargetMap = make(map[string]addressTarget)
	for _, m := range kl.Measurements {
		log.Printf("D! [inputs.KNXListener] group-address mapping for measurement \"%s\"", m.Name)
		for _, ga := range m.Addresses {
			log.Printf("D! [inputs.KNXListener]     %v --> %s", ga, m.Dpt)
			d, err := GetDatapointType(m.Dpt)
			if err != nil {
				return err
			}
			kl.gaTargetMap[ga] = addressTarget{m.Name, d}
		}
	}

	// Connect to the KNX-IP interface
	log.Printf("I! [inputs.KNXListener] Trying to connect to \"%s\" at \"%s\"", kl.ServiceType, kl.ServiceAddress)
	client, err := GetKNXInterface(kl.ServiceType, kl.ServiceAddress)
	if err != nil {
		return err
	}
	kl.client = client
	if kl.ServiceType == "dummy" {
		log.Printf("W! [inputs.KNXListener] Dummy running!")
	} else {
		log.Printf("I! [inputs.KNXListener] Connected!")
	}

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
		// Match GA to DataPointType and measurment name
		ga := msg.Destination.String()
		target, ok := kl.gaTargetMap[ga]
		if ok {
			err := target.datapoint.Unpack(msg.Data)
			if err != nil {
				log.Printf("E! [inputs.KNXListener] Unpacking data failed: %v", err)
				continue
			}
			log.Printf("D! [inputs.KNXListener] Matched GA \"%v\" to measurement \"%v\" with value \"%v\"", ga, target.measurement, target.datapoint)

			// Convert the DatapointValue interface back to its basic type again
			// as otherwise telegraf will not push out the metrics and eat it
			// silently.
			value, err := GetBasicDatapointValue(target.datapoint)
			if err != nil {
				log.Printf("E! [inputs.KNXListener] Type conversion failed: %v", err)
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
			log.Printf("I! [inputs.KNXListener] Ignoring message %+v for unknown GA \"%v\"", msg, ga)
		}
	}
}

func init() {
	inputs.Add("KNXListener", func() telegraf.Input { return &KNXListener{} })
}
