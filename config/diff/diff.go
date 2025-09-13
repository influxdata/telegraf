package diff

import (
	"log"
	"reflect"

	"github.com/fatih/color"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	isnmp "github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs/gnmi"
	"github.com/influxdata/telegraf/plugins/inputs/snmp"
)

// placeholder for the input plugin diff

type InputPluginDiff struct {
	Add []*models.RunningInput
	Del []*models.RunningInput
}

func NewInputPluginDiff() *InputPluginDiff {
	d := &InputPluginDiff{
		Add: make([]*models.RunningInput, 0),
		Del: make([]*models.RunningInput, 0),
	}
	return d
}

func (d *InputPluginDiff) IsEmpty() bool {
	return len(d.Add) == 0 && len(d.Del) == 0
}

func GetPluginUniqueName(input *models.RunningInput) string {
	id := input.ID()
	if id == "" { // this can only hit if plguin in of type pluginWithID and its set to empty
		id = input.Config.ID
	}
	return input.Config.Name + "-" + id
}

func GetPluginNames(op []*models.RunningInput) []string {
	names := make([]string, len(op))
	for i, input := range op {
		names[i] = GetPluginUniqueName(input)
	}
	return names
}

type compareFunc[T any] func(x, y T) bool

var _diffFuncs = map[reflect.Type]func(x, y *telegraf.Input) bool{
	reflect.TypeOf((*snmp.Snmp)(nil)).Elem(): deepEqualSNMP,
	reflect.TypeOf((*gnmi.GNMI)(nil)).Elem(): deepEqualGNMI,
}

// Helper function to compare two slices.
func compareSlices[T any](a, b []T, comp compareFunc[T]) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !comp(a[i], b[i]) {
			return false
		}
	}
	return true
}

func compareSnmpTable(x, y isnmp.Table) bool {
	// specific logic to compare SNMP tables

	if x.Name != y.Name ||
		x.IndexAsTag != y.IndexAsTag ||
		x.Oid != y.Oid ||
		!reflect.DeepEqual(x.InheritTags, y.InheritTags) {
		return false
	}
	return compareSlices(x.Fields, y.Fields, compareSnmpField)
}

func compareSnmpField(x, y isnmp.Field) bool {
	// specific logic to compare SNMP fields
	return x.Name == y.Name &&
		x.Oid == y.Oid &&
		x.OidIndexSuffix == y.OidIndexSuffix &&
		x.OidIndexLength == y.OidIndexLength &&
		x.IsTag == y.IsTag &&
		x.Translate == y.Translate &&
		x.Conversion == y.Conversion &&
		x.SecondaryIndexTable == y.SecondaryIndexTable &&
		x.SecondaryIndexUse == y.SecondaryIndexUse &&
		x.SecondaryOuterJoin == y.SecondaryOuterJoin
}

func compareSecrets(a, b config.Secret) bool {
	// both are empty
	if a.Empty() && b.Empty() {
		return true
	}

	// one is empty
	if a.Empty() || b.Empty() {
		return false
	}

	aVal, aErr := a.Get()

	if aErr != nil {
		return false
	}
	res, err := b.EqualTo(aVal.Bytes())
	if err != nil {
		return false
	}

	return res
}

// CompareClientConfig compares two ClientConfig objects except the Translator field.
func CompareClientConfig(a, b *isnmp.ClientConfig) bool {
	if a.Timeout != b.Timeout ||
		a.Retries != b.Retries ||
		a.Version != b.Version ||
		a.UnconnectedUDPSocket != b.UnconnectedUDPSocket ||
		a.Community != b.Community ||
		a.MaxRepetitions != b.MaxRepetitions ||
		a.ContextName != b.ContextName ||
		a.SecLevel != b.SecLevel ||
		a.SecName != b.SecName ||
		a.AuthProtocol != b.AuthProtocol ||
		a.PrivProtocol != b.PrivProtocol ||
		!compareSlices(a.Path, b.Path, func(x, y string) bool { return x == y }) ||
		!compareSecrets(a.AuthPassword, b.AuthPassword) ||
		!compareSecrets(a.PrivPassword, b.PrivPassword) {
		return false
	}
	return true
}

func deepEqualSNMP(x, y *telegraf.Input) bool {
	// specific logic to compare SNMP inputs
	// with this i can make sure if someone adds same inputs (all fields, tables, etc) more than once i will not consider them as different

	snmpX, okX := (*x).(*snmp.Snmp)
	snmpY, okY := (*y).(*snmp.Snmp)

	if !okX || !okY {
		msg := color.RedString("Type mismatch: x is %T, y is %T\n", *x, *y)
		log.Print("E! [config_dif] ", msg)
		return false
	}

	// Compare addresses, fields, tables, and other properties
	return reflect.DeepEqual(snmpX.Agents, snmpY.Agents) &&
		compareSlices(snmpX.Fields, snmpY.Fields, compareSnmpField) &&
		compareSlices(snmpX.Tables, snmpY.Tables, compareSnmpTable) &&
		CompareClientConfig(&snmpX.ClientConfig, &snmpY.ClientConfig) &&
		snmpX.AgentHostTag == snmpY.AgentHostTag &&
		snmpX.Name == snmpY.Name
}

func deepEqualGNMI(x, y *telegraf.Input) bool {
	// Add specific logic to compare GNMI inputs
	return reflect.DeepEqual(x, y)
}

func Diff(oc, nc []*models.RunningInput) *InputPluginDiff {
	diff := NewInputPluginDiff()

	oldCopy := make([]*models.RunningInput, len(oc))
	copy(oldCopy, oc)

	if len(oc) == 0 {
		diff.Add = nc
		return diff
	}
	if len(nc) == 0 {
		diff.Del = oc
		return diff
	}

	for _, ni := range nc {
		index := -1
		for ind, oi := range oldCopy {
			if reflect.TypeOf(ni.Input).Elem() == reflect.TypeOf(oi.Input).Elem() { // same input type i can compare
				if op, exists := _diffFuncs[reflect.TypeOf(ni.Input).Elem()]; exists {
					if op(&ni.Input, &oi.Input) {
						index = ind
						break
					}
				} else {
					if reflect.DeepEqual(ni.Input, oi.Input) { // same input plugin
						index = ind
						break
					}
				}
			}
		}

		if index == -1 { // i didin't find the input
			diff.Add = append(diff.Add, ni)
		} else { // remove the identified input to speed up next search
			oldCopy = append(oldCopy[:index], oldCopy[index+1:]...)
		}
	}

	// what's left over is either changed or deleted items
	diff.Del = append(diff.Del, oldCopy...)

	return diff
}
