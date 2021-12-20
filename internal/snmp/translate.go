package snmp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
)

// must init, append path for each directory, load module for every file
// or gosmi will fail without saying why
var m sync.Mutex
var once sync.Once
var cache = make(map[string]bool)

func appendPath(path string) {
	m.Lock()
	defer m.Unlock()

	gosmi.AppendPath(path)
}

func loadModule(path string) error {
	m.Lock()
	defer m.Unlock()

	_, err := gosmi.LoadModule(path)
	return err
}

func ClearCache() {
	cache = make(map[string]bool)
}

func LoadMibsFromPath(paths []string, log telegraf.Logger) error {
	once.Do(gosmi.Init)

	for _, mibPath := range paths {
		folders := []string{}

		// Check if we loaded that path already and skip it if so
		m.Lock()
		cached := cache[mibPath]
		cache[mibPath] = true
		m.Unlock()
		if cached {
			continue
		}

		appendPath(mibPath)
		folders = append(folders, mibPath)
		err := filepath.Walk(mibPath, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return fmt.Errorf("no mibs found")
			}
			// symlinks are files so we need to double check if any of them are folders
			// Will check file vs directory later on
			if info.Mode()&os.ModeSymlink != 0 {
				link, err := os.Readlink(path)
				if err != nil {
					log.Warnf("Bad symbolic link %v", link)
				}
				folders = append(folders, link)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Filepath could not be walked: %v", err)
		}

		for _, folder := range folders {
			err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
				// checks if file or directory
				if info.IsDir() {
					appendPath(path)
				} else if info.Mode()&os.ModeSymlink == 0 {
					if err := loadModule(info.Name()); err != nil {
						log.Warn(err)
					}
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("Filepath could not be walked: %v", err)
			}
		}
	}
	return nil
}

// The following is for snmp_trap
type MibEntry struct {
	MibName string
	OidText string
}

func TrapLookup(oid string) (e MibEntry, err error) {
	var node gosmi.SmiNode
	node, err = gosmi.GetNodeByOID(types.OidMustFromString(oid))

	// ensure modules are loaded or node will be empty (might not error)
	if err != nil {
		return e, err
	}

	e.OidText = node.RenderQualified()

	i := strings.Index(e.OidText, "::")
	if i == -1 {
		return e, fmt.Errorf("not found")
	}
	e.MibName = e.OidText[:i]
	e.OidText = e.OidText[i+2:]
	return e, nil
}

// The following is for snmp

func GetIndex(oidNum string, mibPrefix string, node gosmi.SmiNode) (col []string, tagOids map[string]struct{}, err error) {
	// first attempt to get the table's tags
	tagOids = map[string]struct{}{}

	// mimcks grabbing INDEX {} that is returned from snmptranslate -Td MibName
	for _, index := range node.GetIndex() {
		//nolint:staticcheck //assaignment to nil map to keep backwards compatibilty
		tagOids[mibPrefix+index.Name] = struct{}{}
	}

	// grabs all columns from the table
	// mimmicks grabbing everything returned from snmptable -Ch -Cl -c public 127.0.0.1 oidFullName
	col = node.GetRow().AsTable().ColumnOrder

	return col, tagOids, nil
}

//nolint:revive //Too many return variable but necessary
func SnmpTranslateCall(oid string) (mibName string, oidNum string, oidText string, conversion string, node gosmi.SmiNode, err error) {
	var out gosmi.SmiNode
	var end string
	if strings.ContainsAny(oid, "::") {
		// split given oid
		// for example RFC1213-MIB::sysUpTime.0
		s := strings.Split(oid, "::")
		// node becomes sysUpTime.0
		moduleName := s[0]
		_, err := gosmi.GetModule(moduleName)
		if err != nil {
			return oid, oid, oid, oid, gosmi.SmiNode{}, err
		}
		node := s[1]
		if strings.ContainsAny(node, ".") {
			s = strings.Split(node, ".")
			// node becomes sysUpTime
			node = s[0]
			end = "." + s[1]
		}

		out, err = gosmi.GetNode(node)
		if err != nil {
			return oid, oid, oid, oid, out, err
		}

		oidNum = "." + out.RenderNumeric() + end
	} else if strings.ContainsAny(oid, "abcdefghijklnmopqrstuvwxyz") {
		//handle mixed oid ex. .iso.2.3
		s := strings.Split(oid, ".")
		for i := range s {
			if strings.ContainsAny(s[i], "abcdefghijklmnopqrstuvwxyz") {
				out, err = gosmi.GetNode(s[i])
				if err != nil {
					return oid, oid, oid, oid, out, err
				}
				s[i] = out.RenderNumeric()
			}
		}
		oidNum = strings.Join(s, ".")
		out, _ = gosmi.GetNodeByOID(types.OidMustFromString(oidNum))
	} else {
		out, err = gosmi.GetNodeByOID(types.OidMustFromString(oid))
		oidNum = oid
		// ensure modules are loaded or node will be empty (might not error)
		// do not return the err as the oid is numeric and telegraf can continue
		//nolint:nilerr
		if err != nil || out.Name == "iso" {
			return oid, oid, oid, oid, out, nil
		}
	}

	tc := out.GetSubtree()

	for i := range tc {
		// case where the mib doesn't have a conversion so Type struct will be nil
		// prevents seg fault
		if tc[i].Type == nil {
			break
		}
		switch tc[i].Type.Name {
		case "MacAddress", "PhysAddress":
			conversion = "hwaddr"
		case "InetAddressIPv4", "InetAddressIPv6", "InetAddress", "IPSIpAddress":
			conversion = "ipaddr"
		}
	}

	oidText = out.RenderQualified()
	i := strings.Index(oidText, "::")
	if i == -1 {
		return "", oid, oid, oid, out, fmt.Errorf("not found")
	}
	mibName = oidText[:i]
	oidText = oidText[i+2:] + end

	return mibName, oidNum, oidText, conversion, out, nil
}
