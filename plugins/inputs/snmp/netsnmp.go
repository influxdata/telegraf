package snmp

import (
	"bufio"
	"bytes"
	"fmt"
	"log" //nolint:revive
	"os/exec"
	"strings"
	"sync"

	"github.com/influxdata/wlog"
)

//struct that implements the translator interface. This calls existing
//code to exec netsnmp's snmptranslate program
type netsnmpTranslator struct {
}

func NewNetsnmpTranslator() *netsnmpTranslator {
	return &netsnmpTranslator{}
}

type snmpTableCache struct {
	mibName string
	oidNum  string
	oidText string
	fields  []Field
	err     error
}

// execCommand is so tests can mock out exec.Command usage.
var execCommand = exec.Command

// execCmd executes the specified command, returning the STDOUT content.
// If command exits with error status, the output is captured into the returned error.
func execCmd(arg0 string, args ...string) ([]byte, error) {
	if wlog.LogLevel() == wlog.DEBUG {
		quoted := make([]string, 0, len(args))
		for _, arg := range args {
			quoted = append(quoted, fmt.Sprintf("%q", arg))
		}
		log.Printf("D! [inputs.snmp] executing %q %s", arg0, strings.Join(quoted, " "))
	}

	out, err := execCommand(arg0, args...).Output()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s: %w", bytes.TrimRight(err.Stderr, "\r\n"), err)
		}
		return nil, err
	}
	return out, nil
}

var snmpTableCaches map[string]snmpTableCache
var snmpTableCachesLock sync.Mutex

// snmpTable resolves the given OID as a table, providing information about the
// table and fields within.
//nolint:revive
func (n *netsnmpTranslator) SnmpTable(oid string) (
	mibName string, oidNum string, oidText string,
	fields []Field,
	err error) {
	snmpTableCachesLock.Lock()
	if snmpTableCaches == nil {
		snmpTableCaches = map[string]snmpTableCache{}
	}

	var stc snmpTableCache
	var ok bool
	if stc, ok = snmpTableCaches[oid]; !ok {
		stc.mibName, stc.oidNum, stc.oidText, stc.fields, stc.err = n.snmpTableCall(oid)
		snmpTableCaches[oid] = stc
	}

	snmpTableCachesLock.Unlock()
	return stc.mibName, stc.oidNum, stc.oidText, stc.fields, stc.err
}

//nolint:revive
func (n *netsnmpTranslator) snmpTableCall(oid string) (
	mibName string, oidNum string, oidText string,
	fields []Field,
	err error) {
	mibName, oidNum, oidText, _, err = n.SnmpTranslate(oid)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("translating: %w", err)
	}

	mibPrefix := mibName + "::"
	oidFullName := mibPrefix + oidText

	// first attempt to get the table's tags
	tagOids := map[string]struct{}{}
	// We have to guess that the "entry" oid is `oid+".1"`. snmptable and snmptranslate don't seem to have a way to provide the info.
	if out, err := execCmd("snmptranslate", "-Td", oidFullName+".1"); err == nil {
		scanner := bufio.NewScanner(bytes.NewBuffer(out))
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "  INDEX") {
				continue
			}

			i := strings.Index(line, "{ ")
			if i == -1 { // parse error
				continue
			}
			line = line[i+2:]
			i = strings.Index(line, " }")
			if i == -1 { // parse error
				continue
			}
			line = line[:i]
			for _, col := range strings.Split(line, ", ") {
				tagOids[mibPrefix+col] = struct{}{}
			}
		}
	}

	// this won't actually try to run a query. The `-Ch` will just cause it to dump headers.
	out, err := execCmd("snmptable", "-Ch", "-Cl", "-c", "public", "127.0.0.1", oidFullName)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("getting table columns: %w", err)
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	scanner.Scan()
	cols := scanner.Text()
	if len(cols) == 0 {
		return "", "", "", nil, fmt.Errorf("could not find any columns in table")
	}
	for _, col := range strings.Split(cols, " ") {
		if len(col) == 0 {
			continue
		}
		_, isTag := tagOids[mibPrefix+col]
		fields = append(fields, Field{Name: col, Oid: mibPrefix + col, IsTag: isTag})
	}

	return mibName, oidNum, oidText, fields, err
}

type snmpTranslateCache struct {
	mibName    string
	oidNum     string
	oidText    string
	conversion string
	err        error
}

var snmpTranslateCachesLock sync.Mutex
var snmpTranslateCaches map[string]snmpTranslateCache

// snmpTranslate resolves the given OID.
//nolint:revive
func (n *netsnmpTranslator) SnmpTranslate(oid string) (
	mibName string, oidNum string, oidText string,
	conversion string,
	err error) {
	snmpTranslateCachesLock.Lock()
	if snmpTranslateCaches == nil {
		snmpTranslateCaches = map[string]snmpTranslateCache{}
	}

	var stc snmpTranslateCache
	var ok bool
	if stc, ok = snmpTranslateCaches[oid]; !ok {
		// This will result in only one call to snmptranslate running at a time.
		// We could speed it up by putting a lock in snmpTranslateCache and then
		// returning it immediately, and multiple callers would then release the
		// snmpTranslateCachesLock and instead wait on the individual
		// snmpTranslation.Lock to release. But I don't know that the extra complexity
		// is worth it. Especially when it would slam the system pretty hard if lots
		// of lookups are being performed.

		stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.err = snmpTranslateCall(oid)
		snmpTranslateCaches[oid] = stc
	}

	snmpTranslateCachesLock.Unlock()

	return stc.mibName, stc.oidNum, stc.oidText, stc.conversion, stc.err
}

//nolint:revive
func snmpTranslateCall(oid string) (mibName string, oidNum string, oidText string, conversion string, err error) {
	var out []byte
	if strings.ContainsAny(oid, ":abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		out, err = execCmd("snmptranslate", "-Td", "-Ob", oid)
	} else {
		out, err = execCmd("snmptranslate", "-Td", "-Ob", "-m", "all", oid)
		if err, ok := err.(*exec.Error); ok && err.Err == exec.ErrNotFound {
			// Silently discard error if snmptranslate not found and we have a numeric OID.
			// Meaning we can get by without the lookup.
			return "", oid, oid, "", nil
		}
	}
	if err != nil {
		return "", "", "", "", err
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(out))
	ok := scanner.Scan()
	if !ok && scanner.Err() != nil {
		return "", "", "", "", fmt.Errorf("getting OID text: %w", scanner.Err())
	}

	oidText = scanner.Text()

	i := strings.Index(oidText, "::")
	if i == -1 {
		// was not found in MIB.
		if bytes.Contains(out, []byte("[TRUNCATED]")) {
			return "", oid, oid, "", nil
		}
		// not truncated, but not fully found. We still need to parse out numeric OID, so keep going
		oidText = oid
	} else {
		mibName = oidText[:i]
		oidText = oidText[i+2:]
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "  -- TEXTUAL CONVENTION ") {
			tc := strings.TrimPrefix(line, "  -- TEXTUAL CONVENTION ")
			switch tc {
			case "MacAddress", "PhysAddress":
				conversion = "hwaddr"
			case "InetAddressIPv4", "InetAddressIPv6", "InetAddress", "IPSIpAddress":
				conversion = "ipaddr"
			}
		} else if strings.HasPrefix(line, "::= { ") {
			objs := strings.TrimPrefix(line, "::= { ")
			objs = strings.TrimSuffix(objs, " }")

			for _, obj := range strings.Split(objs, " ") {
				if len(obj) == 0 {
					continue
				}
				if i := strings.Index(obj, "("); i != -1 {
					obj = obj[i+1:]
					oidNum += "." + obj[:strings.Index(obj, ")")]
				} else {
					oidNum += "." + obj
				}
			}
			break
		}
	}

	return mibName, oidNum, oidText, conversion, nil
}
