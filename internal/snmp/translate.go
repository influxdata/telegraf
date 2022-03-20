package snmp

import (
	"fmt"
	"io/ioutil"
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

type MibLoader interface {
	// appendPath takes the path of a directory
	appendPath(path string)

	// loadModule takes the name of a file in one of the
	// directories. Basename only, no relative or absolute path
	loadModule(path string) error
}

type GosmiMibLoader struct{}

func (*GosmiMibLoader) appendPath(path string) {
	m.Lock()
	defer m.Unlock()

	gosmi.AppendPath(path)
}

func (*GosmiMibLoader) loadModule(path string) error {
	m.Lock()
	defer m.Unlock()

	_, err := gosmi.LoadModule(path)
	return err
}

func ClearCache() {
	cache = make(map[string]bool)
}

//will give all found folders to gosmi and load in all modules found in the folders
func LoadMibsFromPath(paths []string, log telegraf.Logger, loader MibLoader) error {
	folders, err := walkPaths(paths, log)
	if err != nil {
		return err
	}
	for _, path := range folders {
		loader.appendPath(path)
		modules, err := ioutil.ReadDir(path)
		if err != nil {
			log.Warnf("Can't read directory %v", modules)
		}

		for _, info := range modules {
			if info.Mode()&os.ModeSymlink != 0 {
				symlink := filepath.Join(path, info.Name())
				target, err := filepath.EvalSymlinks(symlink)
				if err != nil {
					log.Warnf("Couldn't evaluate symbolic links for %v: %v", symlink, err)
					continue
				}
				//replace symlink's info with the target's info
				info, err = os.Lstat(target)
				if err != nil {
					log.Warnf("Couldn't stat target %v: %v", target, err)
					continue
				}
			}
			if info.Mode().IsRegular() {
				err := loader.loadModule(info.Name())
				if err != nil {
					log.Warnf("Couldn't load module %v: %v", info.Name(), err)
					continue
				}
			}
		}
	}
	return nil
}

//should walk the paths given and find all folders
func walkPaths(paths []string, log telegraf.Logger) ([]string, error) {
	once.Do(gosmi.Init)
	folders := []string{}

	for _, mibPath := range paths {
		// Check if we loaded that path already and skip it if so
		m.Lock()
		cached := cache[mibPath]
		cache[mibPath] = true
		m.Unlock()
		if cached {
			continue
		}

		err := filepath.Walk(mibPath, func(path string, info os.FileInfo, err error) error {
			if info == nil {
				log.Warnf("No mibs found")
				if os.IsNotExist(err) {
					log.Warnf("MIB path doesn't exist: %q", mibPath)
				} else if err != nil {
					return err
				}
				return nil
			}

			if info.Mode()&os.ModeSymlink != 0 {
				target, err := filepath.EvalSymlinks(path)
				if err != nil {
					log.Warnf("Couldn't evaluate symbolic links for %v: %v", path, err)
				}
				info, err = os.Lstat(target)
				if err != nil {
					log.Warnf("Couldn't stat target %v: %v", target, err)
				}
				path = target
			}
			if info.IsDir() {
				folders = append(folders, path)
			}

			return nil
		})
		if err != nil {
			return folders, fmt.Errorf("Couldn't walk path %q: %v", mibPath, err)
		}
	}
	return folders, nil
}

// The following is for snmp_trap
type MibEntry struct {
	MibName string
	OidText string
}

func TrapLookup(oid string) (e MibEntry, err error) {
	var givenOid types.Oid
	if givenOid, err = types.OidFromString(oid); err != nil {
		return e, fmt.Errorf("could not convert OID %s: %w", oid, err)
	}

	// Get node name
	var node gosmi.SmiNode
	if node, err = gosmi.GetNodeByOID(givenOid); err != nil {
		return e, err
	}
	e.OidText = node.Name

	// Add not found OID part
	if !givenOid.Equals(node.Oid) {
		e.OidText += "." + givenOid[len(node.Oid):].String()
	}

	// Get module name
	module := node.GetModule()
	if module.Name != "<well-known>" {
		e.MibName = module.Name
	}

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
	_, col = node.GetColumns()

	return col, tagOids, nil
}

//nolint:revive //Too many return variable but necessary
func SnmpTranslateCall(oid string) (mibName string, oidNum string, oidText string, conversion string, node gosmi.SmiNode, err error) {
	var out gosmi.SmiNode
	var end string
	if strings.ContainsAny(oid, "::") {
		// split given oid
		// for example RFC1213-MIB::sysUpTime.0
		s := strings.SplitN(oid, "::", 2)
		// moduleName becomes RFC1213
		moduleName := s[0]
		module, err := gosmi.GetModule(moduleName)
		if err != nil {
			return oid, oid, oid, oid, gosmi.SmiNode{}, err
		}
		if s[1] == "" {
			return "", oid, oid, oid, gosmi.SmiNode{}, fmt.Errorf("cannot parse %v\n", oid)
		}
		// node becomes sysUpTime.0
		node := s[1]
		if strings.ContainsAny(node, ".") {
			s = strings.SplitN(node, ".", 2)
			// node becomes sysUpTime
			node = s[0]
			end = "." + s[1]
		}

		out, err = module.GetNode(node)
		if err != nil {
			return oid, oid, oid, oid, out, err
		}

		if oidNum = out.RenderNumeric(); oidNum == "" {
			return oid, oid, oid, oid, out, fmt.Errorf("cannot make %v numeric, please ensure all imported mibs are in the path", oid)
		}

		oidNum = "." + oidNum + end
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
