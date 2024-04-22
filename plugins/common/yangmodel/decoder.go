package yangmodel

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/openconfig/goyang/pkg/yang"
)

var (
	ErrInsufficientData = errors.New("insufficient data")
	ErrNotFound         = errors.New("no such node")
)

type Decoder struct {
	modules *yang.Modules
}

func NewDecoder(paths ...string) (*Decoder, error) {
	modules := yang.NewModules()
	modules.ParseOptions.IgnoreSubmoduleCircularDependencies = true

	var moduleFiles []string
	modulePaths := paths
	unresolved := paths
	for {
		var newlyfound []string
		for _, path := range unresolved {
			entries, err := os.ReadDir(path)
			if err != nil {
				return nil, fmt.Errorf("reading directory %q failed: %w", path, err)
			}
			for _, entry := range entries {
				info, err := entry.Info()
				if err != nil {
					fmt.Printf("Couldn't get info for %q: %v", entry.Name(), err)
					continue
				}

				if info.Mode()&os.ModeSymlink != 0 {
					target, err := filepath.EvalSymlinks(entry.Name())
					if err != nil {
						fmt.Printf("Couldn't evaluate symbolic links for %q: %v", entry.Name(), err)
						continue
					}
					info, err = os.Lstat(target)
					if err != nil {
						fmt.Printf("Couldn't stat target %v: %v", target, err)
						continue
					}
				}

				newPath := filepath.Join(path, info.Name())
				if info.IsDir() {
					newlyfound = append(newlyfound, newPath)
					continue
				}
				if info.Mode().IsRegular() && filepath.Ext(info.Name()) == ".yang" {
					moduleFiles = append(moduleFiles, info.Name())
				}
			}
		}
		if len(newlyfound) == 0 {
			break
		}

		modulePaths = append(modulePaths, newlyfound...)
		unresolved = newlyfound
	}

	// Add the module paths
	modules.AddPath(modulePaths...)
	for _, fn := range moduleFiles {
		if err := modules.Read(fn); err != nil {
			fmt.Printf("reading file %q failed: %v\n", fn, err)
		}
	}
	if errs := modules.Process(); len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &Decoder{modules: modules}, nil
}

func (d *Decoder) FindLeaf(namespace, identifier string) (*yang.Leaf, error) {
	// Get module name from the element
	entry, errs := d.modules.GetModule(namespace)
	if len(errs) > 0 {
		return nil, fmt.Errorf("getting module %q failed: %w", namespace, errors.Join(errs...))
	}

	module, err := d.modules.FindModuleByNamespace(entry.Namespace().NName())
	if err != nil {
		return nil, fmt.Errorf("finding module %q failed: %w", namespace, err)
	}

	for _, grp := range module.Grouping {
		for _, leaf := range grp.Leaf {
			if leaf.Name == identifier {
				return leaf, nil
			}
		}
	}
	return nil, ErrNotFound
}

func DecodeLeaf(leaf *yang.Leaf, value interface{}) (interface{}, error) {
	schema := leaf.Type.YangType

	if schema.Kind != yang.Ybinary {
		return value, nil
	}

	// Binary values are encodes as base64 strings
	s, ok := value.(string)
	if !ok {
		return value, nil
	}

	// Decode the encoded binary values
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return value, err
	}

	switch schema.Name {
	case "ieeefloat32":
		if len(raw) != 4 {
			return raw, fmt.Errorf("%w, expected 4 but got %d bytes", ErrInsufficientData, len(raw))
		}
		return math.Float32frombits(binary.BigEndian.Uint32(raw)), nil
	default:
		return raw, nil
	}
}

func (d *Decoder) DecodeElement(namespace, identifier string, value interface{}) (interface{}, error) {
	leaf, err := d.FindLeaf(namespace, identifier)
	if err != nil {
		return nil, fmt.Errorf("finding %s failed: %w", identifier, err)
	}

	return DecodeLeaf(leaf, value)
}
