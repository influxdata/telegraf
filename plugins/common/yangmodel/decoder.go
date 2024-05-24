package yangmodel

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"github.com/openconfig/goyang/pkg/yang"
)

var (
	ErrInsufficientData = errors.New("insufficient data")
	ErrNotFound         = errors.New("no such node")
)

type Decoder struct {
	modules   map[string]*yang.Module
	rootNodes map[string][]yang.Node
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

	// Get all root nodes defined in models with their origin. We require
	// those nodes to later resolve paths to YANG model leaf nodes...
	moduleLUT := make(map[string]*yang.Module)
	moduleRootNodes := make(map[string][]yang.Node)
	for _, m := range modules.Modules {
		// Check if we processed the module already
		if _, found := moduleLUT[m.Name]; found {
			continue
		}
		// Create a module mapping for easily finding modules by name
		moduleLUT[m.Name] = m

		// Determine the origin defined in the module
		var prefix string
		for _, imp := range m.Import {
			if imp.Name == "openconfig-extensions" {
				prefix = imp.Name
				if imp.Prefix != nil {
					prefix = imp.Prefix.Name
				}
				break
			}
		}

		var moduleOrigin string
		if prefix != "" {
			for _, e := range m.Extensions {
				if e.Keyword == prefix+":origin" || e.Keyword == "origin" {
					moduleOrigin = e.Argument
					break
				}
			}
		}
		for _, u := range m.Uses {
			root, err := yang.FindNode(m, u.Name)
			if err != nil {
				return nil, err
			}
			moduleRootNodes[moduleOrigin] = append(moduleRootNodes[moduleOrigin], root)
		}
	}

	return &Decoder{modules: moduleLUT, rootNodes: moduleRootNodes}, nil
}

func (d *Decoder) FindLeaf(name, identifier string) (*yang.Leaf, error) {
	// Get module name from the element
	module, found := d.modules[name]
	if !found {
		return nil, fmt.Errorf("cannot find module %q", name)
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

func DecodeLeafValue(leaf *yang.Leaf, value interface{}) (interface{}, error) {
	schema := leaf.Type.YangType

	// Ignore all non-string values as the types seem already converted...
	s, ok := value.(string)
	if !ok {
		return value, nil
	}

	switch schema.Kind {
	case yang.Ybinary:
		// Binary values are encodes as base64 string, so decode the string
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
	case yang.Yint8:
		v, err := strconv.ParseInt(s, 10, 8)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return int8(v), nil
	case yang.Yint16:
		v, err := strconv.ParseInt(s, 10, 16)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return int16(v), nil
	case yang.Yint32:
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return int32(v), nil
	case yang.Yint64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return v, nil
	case yang.Yuint8:
		v, err := strconv.ParseUint(s, 10, 8)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return uint8(v), nil
	case yang.Yuint16:
		v, err := strconv.ParseUint(s, 10, 16)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return uint16(v), nil
	case yang.Yuint32:
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return uint32(v), nil
	case yang.Yuint64:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return v, nil
	case yang.Ydecimal64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return value, fmt.Errorf("parsing %s %q failed: %w", yang.TypeKindToName[schema.Kind], s, err)
		}
		return v, nil
	}
	return value, nil
}

func (d *Decoder) DecodeLeafElement(namespace, identifier string, value interface{}) (interface{}, error) {
	leaf, err := d.FindLeaf(namespace, identifier)
	if err != nil {
		return nil, fmt.Errorf("finding %s failed: %w", identifier, err)
	}

	return DecodeLeafValue(leaf, value)
}

func (d *Decoder) DecodePathElement(origin, path string, value interface{}) (interface{}, error) {
	rootNodes, found := d.rootNodes[origin]
	if !found || len(rootNodes) == 0 {
		return value, nil
	}

	for _, root := range rootNodes {
		node, _ := yang.FindNode(root, path)
		if node == nil {
			// The path does not exist in this root node
			continue
		}
		// We do expect a leaf node...
		if leaf, ok := node.(*yang.Leaf); ok {
			return DecodeLeafValue(leaf, value)
		}
	}

	return value, nil
}
