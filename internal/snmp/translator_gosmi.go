package snmp

import (
	"fmt"

	"github.com/sleepinggenius2/gosmi/models"

	"github.com/influxdata/telegraf"
)

type gosmiTranslator struct {
}

func NewGosmiTranslator(paths []string, log telegraf.Logger) (*gosmiTranslator, error) {
	err := LoadMibsFromPath(paths, log, &GosmiMibLoader{})
	if err == nil {
		return &gosmiTranslator{}, nil
	}
	return nil, err
}

//nolint:revive //function-result-limit conditionally 5 return results allowed
func (g *gosmiTranslator) SnmpTranslate(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
	mibName, oidNum, oidText, conversion, _, err = SnmpTranslateCall(oid)
	return mibName, oidNum, oidText, conversion, err
}

// snmpTable resolves the given OID as a table, providing information about the
// table and fields within.
//
//nolint:revive //Too many return variable but necessary
func (g *gosmiTranslator) SnmpTable(oid string) (
	mibName string, oidNum string, oidText string,
	fields []Field,
	err error) {
	mibName, oidNum, oidText, _, node, err := SnmpTranslateCall(oid)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("translating: %w", err)
	}

	mibPrefix := mibName + "::"

	col, tagOids := GetIndex(mibPrefix, node)
	for _, c := range col {
		_, isTag := tagOids[mibPrefix+c]
		fields = append(fields, Field{Name: c, Oid: mibPrefix + c, IsTag: isTag})
	}

	return mibName, oidNum, oidText, fields, nil
}

func (g *gosmiTranslator) SnmpFormatEnum(oid string, value interface{}, full bool) (string, error) {
	//nolint:dogsled // only need to get the node
	_, _, _, _, node, err := SnmpTranslateCall(oid)

	if err != nil {
		return "", err
	}

	var v models.Value
	if full {
		v = node.FormatValue(value, models.FormatEnumName, models.FormatEnumValue)
	} else {
		v = node.FormatValue(value, models.FormatEnumName)
	}

	return v.Formatted, nil
}
