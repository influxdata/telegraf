package snmp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/sleepinggenius2/gosmi"
	"github.com/sleepinggenius2/gosmi/types"
)

type MibEntry struct {
	MibName string
	OidText string
}

// must init, append path for each directory, load module for every file
// or gosmi will fail without saying why
func GetMibsPath(paths []string, log telegraf.Logger) error {
	gosmi.Init()
	var folders []string
	for _, mibPath := range paths {
		gosmi.AppendPath(mibPath)
		folders = append(folders, mibPath)
		err := filepath.Walk(mibPath, func(path string, info os.FileInfo, err error) error {
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
			return fmt.Errorf("Filepath could not be walked %v", err)
		}
		for _, folder := range folders {
			err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
				// checks if file or directory
				if info.IsDir() {
					gosmi.AppendPath(path)
				} else if info.Mode()&os.ModeSymlink == 0 {
					_, err := gosmi.LoadModule(info.Name())
					if err != nil {
						log.Warnf("Module could not be loaded %v", err)
					}
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("Filepath could not be walked %v", err)
			}
		}
		folders = []string{}
	}
	return nil
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
