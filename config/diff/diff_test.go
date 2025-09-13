package diff

import (
	"reflect"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	isnmp "github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/inputs/gnmi"
	"github.com/influxdata/telegraf/plugins/inputs/snmp"
)

func TestGetPluginName(t *testing.T) {
	input := &models.RunningInput{
		Config: &models.InputConfig{
			Name: "example",
			ID:   "1234",
		},
	}

	expected := "example-1234"
	actual := GetPluginUniqueName(input)

	if actual != expected {
		t.Errorf("GetPluginUniqueName() = %v, want %v", actual, expected)
	}
}

func TestGetPluginNames(t *testing.T) {
	inputs := []*models.RunningInput{
		{
			Config: &models.InputConfig{
				Name: "example1",
				ID:   "1234",
			},
		},
		{
			Config: &models.InputConfig{
				Name: "example2",
				ID:   "5678",
			},
		},
	}

	expected := []string{"example1-1234", "example2-5678"}
	actual := GetPluginNames(inputs)

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("GetPluginNames() = %v, want %v", actual, expected)
	}
}

func TestCompareSlices(t *testing.T) {
	intComp := func(x, y int) bool { return x == y }
	stringComp := func(x, y string) bool { return x == y }

	intSlice1 := []int{1, 2, 3}
	intSlice2 := []int{1, 2, 3}
	intSlice3 := []int{1, 2, 4}

	stringSlice1 := []string{"a", "b", "c"}
	stringSlice2 := []string{"a", "b", "c"}
	stringSlice3 := []string{"a", "b", "d"}

	if !compareSlices(intSlice1, intSlice2, intComp) {
		t.Errorf("compareSlices(intSlice1, intSlice2) = false, want true")
	}

	if compareSlices(intSlice1, intSlice3, intComp) {
		t.Errorf("compareSlices(intSlice1, intSlice3) = true, want false")
	}

	if !compareSlices(stringSlice1, stringSlice2, stringComp) {
		t.Errorf("compareSlices(stringSlice1, stringSlice2) = false, want true")
	}

	if compareSlices(stringSlice1, stringSlice3, stringComp) {
		t.Errorf("compareSlices(stringSlice1, stringSlice3) = true, want false")
	}
}

func TestDiffFuncs(t *testing.T) {
	snmp1 := &models.RunningInput{
		Input: &snmp.Snmp{},
	}
	snmp2 := &models.RunningInput{
		Input: &snmp.Snmp{},
	}
	gnmi1 := &models.RunningInput{
		Input: &gnmi.GNMI{},
	}
	gnmi2 := &models.RunningInput{
		Input: &gnmi.GNMI{},
	}

	if !_diffFuncs[reflect.TypeOf(snmp1.Input).Elem()](&snmp1.Input, &snmp2.Input) {
		t.Errorf("deepEqualSNMP() = false, want true")
	}

	if _diffFuncs[reflect.TypeOf(snmp1.Input).Elem()](&snmp1.Input, &gnmi1.Input) {
		t.Errorf("deepEqualSNMP() = true, want false")
	}

	if _diffFuncs[reflect.TypeOf(gnmi1.Input).Elem()](&snmp1.Input, &gnmi1.Input) {
		t.Errorf("deepEqualGNMI() = true, want false")
	}
	if !_diffFuncs[reflect.TypeOf(gnmi2.Input).Elem()](&gnmi2.Input, &gnmi1.Input) {
		t.Errorf("deepEqualGNMI() = false, want true")
	}
}

func TestDiff(t *testing.T) {
	tests := map[string]struct {
		input1       []*models.RunningInput
		input2       []*models.RunningInput
		expectedDiff InputPluginDiff
	}{
		"same SNMP inputs": {
			input1: []*models.RunningInput{
				{Input: &snmp.Snmp{}},
			},
			input2: []*models.RunningInput{
				{Input: &snmp.Snmp{}},
			},
			expectedDiff: InputPluginDiff{
				Add: make([]*models.RunningInput, 0),
				Del: make([]*models.RunningInput, 0),
			},
		},
		"SNMP to GNMI": {
			input1: []*models.RunningInput{
				{Input: &snmp.Snmp{}},
			},
			input2: []*models.RunningInput{
				{Input: &gnmi.GNMI{}},
			},
			expectedDiff: InputPluginDiff{
				Add: []*models.RunningInput{
					{Input: &gnmi.GNMI{}},
				},
				Del: []*models.RunningInput{
					{Input: &snmp.Snmp{}},
				},
			},
		},
		"same GNMI inputs": {
			input1: []*models.RunningInput{
				{Input: &gnmi.GNMI{}},
			},
			input2: []*models.RunningInput{
				{Input: &gnmi.GNMI{}},
			},
			expectedDiff: InputPluginDiff{
				Add: make([]*models.RunningInput, 0),
				Del: make([]*models.RunningInput, 0),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			d := Diff(tc.input1, tc.input2)
			if !reflect.DeepEqual(d, &tc.expectedDiff) {
				t.Errorf("Diff(%v, %v) = %v, want %v", tc.input1, tc.input2, d, tc.expectedDiff)
			}
		})
	}
}

func TestCompareSnmpField(t *testing.T) {
	tests := map[string]struct {
		field1      isnmp.Field
		field2      isnmp.Field
		expectedRes bool
	}{
		"equal fields": {
			field1: isnmp.Field{
				Name:                "field1",
				Oid:                 "1.3.6.1.2.1.1.1",
				OidIndexSuffix:      "0",
				OidIndexLength:      1,
				IsTag:               true,
				Translate:           true,
				SecondaryIndexTable: true,
				SecondaryIndexUse:   true,
				SecondaryOuterJoin:  true,
			},
			field2: isnmp.Field{
				Name:                "field1",
				Oid:                 "1.3.6.1.2.1.1.1",
				OidIndexSuffix:      "0",
				OidIndexLength:      1,
				IsTag:               true,
				Translate:           true,
				SecondaryIndexTable: true,
				SecondaryIndexUse:   true,
				SecondaryOuterJoin:  true,
			},
			expectedRes: true,
		},
		"equal fields case 2": {
			field1: isnmp.Field{
				Name: "field_random",
				Oid:  ".1.3.6.1.2.4.6.7",
			},
			field2: isnmp.Field{
				Name: "field_random",
				Oid:  ".1.3.6.1.2.4.6.7",
			},
			expectedRes: true,
		},
		"equal fields case 3": {
			field1: isnmp.Field{
				Name:  "field_random_tag",
				Oid:   ".1.3.6.1.2.4.6.7",
				IsTag: true,
			},
			field2: isnmp.Field{
				Name:  "field_random_tag",
				Oid:   ".1.3.6.1.2.4.6.7",
				IsTag: true,
			},
			expectedRes: true,
		},
		"field with same oid not a tag": {
			field1: isnmp.Field{
				Name:  "field_random_tag",
				Oid:   ".1.3.6.1.2.4.6.7",
				IsTag: true,
			},
			field2: isnmp.Field{
				Name:  "field_random_tag",
				Oid:   ".1.3.6.1.2.4.6.7",
				IsTag: false,
			},
			expectedRes: false,
		},
		"field with different name": {
			field1: isnmp.Field{
				Name:                "field1",
				Oid:                 "1.3.6.1.2.1.1.1",
				OidIndexSuffix:      "0",
				OidIndexLength:      1,
				IsTag:               true,
				Translate:           true,
				SecondaryIndexTable: true,
				SecondaryIndexUse:   true,
				SecondaryOuterJoin:  true,
			},
			field2: isnmp.Field{
				Name:                "field2",
				Oid:                 "1.3.6.1.2.1.1.1",
				OidIndexSuffix:      "0",
				OidIndexLength:      1,
				IsTag:               true,
				Translate:           true,
				SecondaryIndexTable: true,
				SecondaryIndexUse:   true,
				SecondaryOuterJoin:  true,
			},
			expectedRes: false,
		},
		"different oid": {
			field1: isnmp.Field{
				Name:           "field1",
				Oid:            "1.3.6.1.2.1.1.1",
				OidIndexSuffix: "0",
				OidIndexLength: 1,
				IsTag:          true,
			},
			field2: isnmp.Field{
				Name:           "field1",
				Oid:            "1.3.6.1.2.1.1.2",
				OidIndexSuffix: "0",
				OidIndexLength: 1,
				IsTag:          true,
			},
			expectedRes: false,
		},
		"different tag": {
			field1: isnmp.Field{
				Name:                "field1",
				Oid:                 "1.3.6.1.2.1.1.1",
				OidIndexSuffix:      "0",
				OidIndexLength:      1,
				IsTag:               true,
				Translate:           true,
				SecondaryIndexTable: true,
				SecondaryIndexUse:   true,
				SecondaryOuterJoin:  true,
			},
			field2: isnmp.Field{
				Name:                "field1",
				Oid:                 "1.3.6.1.2.1.1.1",
				OidIndexSuffix:      "0",
				OidIndexLength:      1,
				IsTag:               false,
				Translate:           true,
				SecondaryIndexTable: true,
				SecondaryIndexUse:   true,
				SecondaryOuterJoin:  true,
			},
			expectedRes: false,
		},
		"different conversion": {
			field1: isnmp.Field{
				Name:       "field1",
				Oid:        "1.3.6.1.2.1.1.1",
				Conversion: "conversion1",
			},
			field2: isnmp.Field{
				Name:       "field1",
				Oid:        "1.3.6.1.2.1.1.1",
				Conversion: "conversion2",
			},
			expectedRes: false,
		},
		"same conversion": {
			field1: isnmp.Field{
				Name:       "field1",
				Oid:        "1.3.6.1.2.1.1.1",
				Conversion: "conversion1",
			},
			field2: isnmp.Field{
				Name:       "field1",
				Oid:        "1.3.6.1.2.1.1.1",
				Conversion: "conversion1",
			},
			expectedRes: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if equal := compareSnmpField(tc.field1, tc.field2); equal != tc.expectedRes {
				t.Errorf("compareSnmpField(%v, %v) = %v, want %v", tc.field1, tc.field2, equal, tc.expectedRes)
			}
		})
	}
}

func TestCompareSnmpTable(t *testing.T) {
	tests := map[string]struct {
		table1      isnmp.Table
		table2      isnmp.Table
		expectedRes bool
	}{
		"equal tables": {
			table1: isnmp.Table{
				Name:        "table1",
				IndexAsTag:  true,
				Oid:         "1.3.6.1.2.1.2.2",
				InheritTags: []string{"tag1", "tag2"},
				Fields: []isnmp.Field{
					{Name: "field1", Oid: "1.3.6.1.2.1.2.2.1.1"},
					{Name: "field2", Oid: "1.3.6.1.2.1.2.2.1.2"},
				},
			},
			table2: isnmp.Table{
				Name:        "table1",
				IndexAsTag:  true,
				Oid:         "1.3.6.1.2.1.2.2",
				InheritTags: []string{"tag1", "tag2"},
				Fields: []isnmp.Field{
					{Name: "field1", Oid: "1.3.6.1.2.1.2.2.1.1"},
					{Name: "field2", Oid: "1.3.6.1.2.1.2.2.1.2"},
				},
			},
			expectedRes: true,
		},
		"same table with different names": {
			table1: isnmp.Table{
				Name:        "table1",
				IndexAsTag:  true,
				Oid:         "1.3.6.1.2.1.2.2",
				InheritTags: []string{"tag1", "tag2"},
				Fields: []isnmp.Field{
					{Name: "field1", Oid: "1.3.6.1.2.1.2.2.1.1"},
					{Name: "field2", Oid: "1.3.6.1.2.1.2.2.1.2"},
				},
			},
			table2: isnmp.Table{
				Name:        "table2",
				IndexAsTag:  true,
				Oid:         "1.3.6.1.2.1.2.2",
				InheritTags: []string{"tag1", "tag2"},
				Fields: []isnmp.Field{
					{Name: "field1", Oid: "1.3.6.1.2.1.2.2.1.1"},
					{Name: "field2", Oid: "1.3.6.1.2.1.2.2.1.2"},
				},
			},
			expectedRes: false,
		},
		"different fields": {
			table1: isnmp.Table{
				Name:        "table1",
				IndexAsTag:  true,
				Oid:         "1.3.6.1.2.1.2.2",
				InheritTags: []string{"tag1", "tag2"},
				Fields: []isnmp.Field{
					{Name: "field1", Oid: "1.3.6.1.2.1.2.2.1.1"},
					{Name: "field2", Oid: "1.3.6.1.2.1.2.2.1.2"},
				},
			},
			table2: isnmp.Table{
				Name:        "table1",
				IndexAsTag:  true,
				Oid:         "1.3.6.1.2.1.2.2",
				InheritTags: []string{"tag1", "tag2"},
				Fields: []isnmp.Field{
					{Name: "field3", Oid: "1.3.6.1.2.1.2.2.1.3"},
				},
			},
			expectedRes: false,
		},
		"different OIDs": {
			table1: isnmp.Table{
				Name:        "table1",
				IndexAsTag:  true,
				Oid:         "1.3.6.1.2.1.2.2",
				InheritTags: []string{"tag1", "tag2"},
				Fields: []isnmp.Field{
					{Name: "field1", Oid: "1.3.6.1.2.1.2.2.1.1"},
					{Name: "field2", Oid: "1.3.6.1.2.1.2.2.1.2"},
				},
			},
			table2: isnmp.Table{
				Name:        "table1",
				IndexAsTag:  true,
				Oid:         "1.3.6.1.2.1.2.3",
				InheritTags: []string{"tag1", "tag2"},
				Fields: []isnmp.Field{
					{Name: "field1", Oid: "1.3.6.1.2.1.2.2.1.1"},
					{Name: "field2", Oid: "1.3.6.1.2.1.2.2.1.2"},
				},
			},
			expectedRes: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if equal := compareSnmpTable(tc.table1, tc.table2); equal != tc.expectedRes {
				t.Errorf("compareSnmpTable(%v, %v) = %v, want %v", tc.table1, tc.table2, equal, tc.expectedRes)
			}
		})
	}
}

func TestCompareSecrets(t *testing.T) {
	tests := map[string]struct {
		secret1     config.Secret
		secret2     config.Secret
		expectedRes bool
	}{
		"both secrets empty": {
			secret1:     config.Secret{},
			secret2:     config.Secret{},
			expectedRes: true,
		},
		"one secret empty": {
			secret1:     config.Secret{},
			secret2:     config.NewSecret([]byte("secret")),
			expectedRes: false,
		},
		"equal secrets": {
			secret1:     config.NewSecret([]byte("secret")),
			secret2:     config.NewSecret([]byte("secret")),
			expectedRes: true,
		},
		"different secrets": {
			secret1:     config.NewSecret([]byte("secret1")),
			secret2:     config.NewSecret([]byte("secret2")),
			expectedRes: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if equal := compareSecrets(tc.secret1, tc.secret2); equal != tc.expectedRes {
				t.Errorf("compareSecrets(%v, %v) = %v, want %v", tc.secret1, tc.secret2, equal, tc.expectedRes)
			}
		})
	}
}

func TestCompareClientConfig(t *testing.T) {
	tests := map[string]struct {
		config1     isnmp.ClientConfig
		config2     isnmp.ClientConfig
		expectedRes bool
	}{
		"equal configs": {
			config1: isnmp.ClientConfig{
				Timeout:              5,
				Retries:              3,
				Version:              2,
				UnconnectedUDPSocket: true,
				Community:            "public",
				MaxRepetitions:       10,
				ContextName:          "context",
				SecLevel:             "authPriv",
				SecName:              "user",
				AuthProtocol:         "MD5",
				PrivProtocol:         "DES",
				Path:                 []string{"path1", "path2"},
				AuthPassword:         config.NewSecret([]byte("authpass")),
				PrivPassword:         config.NewSecret([]byte("privpass")),
			},
			config2: isnmp.ClientConfig{
				Timeout:              5,
				Retries:              3,
				Version:              2,
				UnconnectedUDPSocket: true,
				Community:            "public",
				MaxRepetitions:       10,
				ContextName:          "context",
				SecLevel:             "authPriv",
				SecName:              "user",
				AuthProtocol:         "MD5",
				PrivProtocol:         "DES",
				Path:                 []string{"path1", "path2"},
				AuthPassword:         config.NewSecret([]byte("authpass")),
				PrivPassword:         config.NewSecret([]byte("privpass")),
			},
			expectedRes: true,
		},
		"different timeouts": {
			config1:     isnmp.ClientConfig{Timeout: 5},
			config2:     isnmp.ClientConfig{Timeout: 10},
			expectedRes: false,
		},
		"different paths": {
			config1:     isnmp.ClientConfig{Path: []string{"path1"}},
			config2:     isnmp.ClientConfig{Path: []string{"path2"}},
			expectedRes: false,
		},
		"different secrets": {
			config1:     isnmp.ClientConfig{AuthPassword: config.NewSecret([]byte("authpass1"))},
			config2:     isnmp.ClientConfig{AuthPassword: config.NewSecret([]byte("authpass2"))},
			expectedRes: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if equal := CompareClientConfig(&tc.config1, &tc.config2); equal != tc.expectedRes {
				t.Errorf("CompareClientConfig(%v, %v) = %v, want %v", tc.config1, tc.config2, equal, tc.expectedRes)
			}
		})
	}
}

func TestDeepEqualSNMP(t *testing.T) {
	tests := map[string]struct {
		snmp1       *snmp.Snmp
		snmp2       *snmp.Snmp
		expectedRes bool
	}{
		"equal SNMP configs": {
			snmp1: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			snmp2: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			expectedRes: true,
		},
		"equal SNMP configs with different id": {
			snmp1: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
				PluginID:     "1234",
			},
			snmp2: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
				PluginID:     "5678",
			},
			expectedRes: true,
		},
		"different agents": {
			snmp1: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			snmp2: &snmp.Snmp{
				Agents:       []string{"agent3"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			expectedRes: false,
		},
		"different fields": {
			snmp1: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			snmp2: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field2", Oid: "1.3.6.1.2.1.1.2"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			expectedRes: false,
		},
		"different tables": {
			snmp1: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			snmp2: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table2", Oid: "1.3.6.1.2.1.2.3"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			expectedRes: false,
		},
		"different client config": {
			snmp1: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 5},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			snmp2: &snmp.Snmp{
				Agents:       []string{"agent1", "agent2"},
				Fields:       []isnmp.Field{{Name: "field1", Oid: "1.3.6.1.2.1.1.1"}},
				Tables:       []isnmp.Table{{Name: "table1", Oid: "1.3.6.1.2.1.2.2"}},
				ClientConfig: isnmp.ClientConfig{Timeout: 10},
				AgentHostTag: "host1",
				Name:         "snmp1",
			},
			expectedRes: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			snmpInput1 := telegraf.Input(tc.snmp1)
			snmpInput2 := telegraf.Input(tc.snmp2)
			if equal := deepEqualSNMP(&snmpInput1, &snmpInput2); equal != tc.expectedRes {
				t.Errorf("deepEqualSNMP(%v, %v) = %v, want %v", tc.snmp1, tc.snmp2, equal, tc.expectedRes)
			}
		})
	}

	snmpInput := telegraf.Input(&snmp.Snmp{})
	gnmiInput := telegraf.Input(&gnmi.GNMI{})
	res := deepEqualSNMP(&snmpInput, &gnmiInput)
	t.Run("TestDeepEqualSNMPWithDifferentTypes", func(t *testing.T) {
		if res {
			t.Errorf("Expected res to be false, got true")
		}
	})
}
