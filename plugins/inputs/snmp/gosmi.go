package snmp

import (
	"fmt"
	"sync"

	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/models"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/snmp"
)

type gosmiTranslator struct {
}

func NewGosmiTranslator(paths []string, log telegraf.Logger) (*gosmiTranslator, error) {
	err := snmp.LoadMibsFromPath(paths, log, &snmp.GosmiMibLoader{})
	if err == nil {
		return &gosmiTranslator{}, nil
	}
	return nil, err
}

type gosmiSnmpTranslateCache struct {
	mibName    string
	oidNum     string
	oidText    string
	conversion string
	node       gosmi.SmiNode
	err        error
}

var gosmiSnmpTranslateCachesLock sync.Mutex
var gosmiSnmpTranslateCaches map[string]gosmiSnmpTranslateCache

//nolint:revive //function-result-limit conditionally 5 return results allowed
func (g *gosmiTranslator) SnmpTranslate(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
	mibName, oidNum, oidText, conversion, _, err = g.SnmpTranslateFull(oid)
	return mibName, oidNum, oidText, conversion, err
}

//nolint:revive //function-result-limit conditionally 6 return results allowed
func (g *gosmiTranslator) SnmpTranslateFull(oid string) (
	mibName string, oidNum string, oidText string,
	conversion string,
	node gosmi.SmiNode,
	err error) {
	gosmiSnmpTranslateCachesLock.Lock()
	if gosmiSnmpTranslateCaches == nil {
		gosmiSnmpTranslateCaches = map[string]gosmiSnmpTranslateCache{}
	}

	var stc gosmiSnmpTranslateCache
	var ok bool
	if stc, ok = gosmiSnmpTranslateCaches[oid]; !ok {
		// This will result in only one call to snmptranslate running at a time.
		// We could speed it up by putting a lock in snmpTranslateCache and then
		// returning it immediately, and multiple callers would then release the
		// snmpTranslateCachesLock and instead wait on the individual
		// snmpTranslation.Lock to release. But I don't know that the extra complexity
		// is worth it. Especially when it would slam the system pretty hard if lots
		// of lookups are being performed.

		stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.node, stc.err = snmp.SnmpTranslateCall(oid)
		gosmiSnmpTranslateCaches[oid] = stc
	}

	gosmiSnmpTranslateCachesLock.Unlock()

	return stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.node, stc.err
}

type gosmiSnmpTableCache struct {
	mibName string
	oidNum  string
	oidText string
	fields  []Field
	err     error
}

var gosmiSnmpTableCaches map[string]gosmiSnmpTableCache
var gosmiSnmpTableCachesLock sync.Mutex

// snmpTable resolves the given OID as a table, providing information about the
// table and fields within.
//
//nolint:revive //Too many return variable but necessary
func (g *gosmiTranslator) SnmpTable(oid string) (
	mibName string, oidNum string, oidText string,
	fields []Field,
	err error) {
	gosmiSnmpTableCachesLock.Lock()
	if gosmiSnmpTableCaches == nil {
		gosmiSnmpTableCaches = map[string]gosmiSnmpTableCache{}
	}

	var stc gosmiSnmpTableCache
	var ok bool
	if stc, ok = gosmiSnmpTableCaches[oid]; !ok {
		stc.mibName, stc.oidNum, stc.oidText, stc.fields, stc.err = g.SnmpTableCall(oid)
		gosmiSnmpTableCaches[oid] = stc
	}

	gosmiSnmpTableCachesLock.Unlock()
	return stc.mibName, stc.oidNum, stc.oidText, stc.fields, stc.err
}

//nolint:revive //Too many return variable but necessary
func (g *gosmiTranslator) SnmpTableCall(oid string) (mibName string, oidNum string, oidText string, fields []Field, err error) {
	mibName, oidNum, oidText, _, node, err := g.SnmpTranslateFull(oid)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("translating: %w", err)
	}

	mibPrefix := mibName + "::"

	col, tagOids := snmp.GetIndex(mibPrefix, node)
	for _, c := range col {
		_, isTag := tagOids[mibPrefix+c]
		fields = append(fields, Field{Name: c, Oid: mibPrefix + c, IsTag: isTag})
	}

	return mibName, oidNum, oidText, fields, nil
}

func (g *gosmiTranslator) SnmpFormatEnum(oid string, value interface{}, full bool) (string, error) {
	//nolint:dogsled // only need to get the node
	_, _, _, _, node, err := g.SnmpTranslateFull(oid)

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
