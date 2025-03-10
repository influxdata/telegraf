package xpath

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	path "github.com/antchfx/xpath"
	"github.com/bufbuild/protocompile"
	"github.com/srebhan/protobufquery"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/influxdata/telegraf"
)

type protobufDocument struct {
	MessageFiles []string
	MessageType  string
	ImportPaths  []string
	SkipBytes    int64
	Log          telegraf.Logger

	msg          *dynamicpb.Message
	unmarshaller proto.UnmarshalOptions
}

func (d *protobufDocument) Init() error {
	// Check the message definition and type
	if len(d.MessageFiles) == 0 {
		return errors.New("protocol-buffer files not set")
	}
	if d.MessageType == "" {
		return errors.New("protocol-buffer message-type not set")
	}

	// Load the file descriptors from the given protocol-buffer definition
	ctx := context.Background()
	resolver := &protocompile.SourceResolver{ImportPaths: d.ImportPaths}
	compiler := &protocompile.Compiler{
		Resolver: protocompile.WithStandardImports(resolver),
	}
	files, err := compiler.Compile(ctx, d.MessageFiles...)
	if err != nil {
		return fmt.Errorf("parsing protocol-buffer definition failed: %w", err)
	}
	if len(files) < 1 {
		return errors.New("files do not contain a file descriptor")
	}

	// Register all definitions in the file in the global registry
	var registry protoregistry.Files
	for _, f := range files {
		if err := registry.RegisterFile(f); err != nil {
			return fmt.Errorf("adding file %q to registry failed: %w", f.Path(), err)
		}
	}

	d.unmarshaller = proto.UnmarshalOptions{
		RecursionLimit: protowire.DefaultRecursionLimit,
		Resolver:       dynamicpb.NewTypes(&registry),
	}

	// Lookup given type in the loaded file descriptors
	msgFullName := protoreflect.FullName(d.MessageType)
	descriptor, err := registry.FindDescriptorByName(msgFullName)
	if err != nil {
		d.Log.Infof("Could not find %q... Known messages:", msgFullName)

		var known []string
		registry.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
			name := strings.TrimSpace(string(fd.FullName()))
			if name != "" {
				known = append(known, name)
			}
			return true
		})
		sort.Strings(known)
		for _, name := range known {
			d.Log.Infof("  %s", name)
		}
		return err
	}

	// Get a prototypical message for later use
	msgDesc, ok := descriptor.(protoreflect.MessageDescriptor)
	if !ok {
		return fmt.Errorf("%q is not a message descriptor (%T)", msgFullName, descriptor)
	}

	d.msg = dynamicpb.NewMessage(msgDesc)
	if d.msg == nil {
		return fmt.Errorf("creating message template for %q failed", msgDesc.FullName())
	}

	return nil
}

func (d *protobufDocument) Parse(buf []byte) (dataNode, error) {
	msg := d.msg.New()

	// Unmarshal the received buffer
	if err := d.unmarshaller.Unmarshal(buf[d.SkipBytes:], msg.Interface()); err != nil {
		hexbuf := hex.EncodeToString(buf)
		d.Log.Debugf("raw data (hex): %q (skip %d bytes)", hexbuf, d.SkipBytes)
		return nil, err
	}

	return protobufquery.Parse(msg)
}

func (*protobufDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
	// If this panics it's a programming error as we changed the document type while processing
	native, err := protobufquery.QueryAll(node.(*protobufquery.Node), expr)
	if err != nil {
		return nil, err
	}

	nodes := make([]dataNode, 0, len(native))
	for _, n := range native {
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (*protobufDocument) CreateXPathNavigator(node dataNode) path.NodeNavigator {
	// If this panics it's a programming error as we changed the document type while processing
	return protobufquery.CreateXPathNavigator(node.(*protobufquery.Node))
}

func (d *protobufDocument) GetNodePath(node, relativeTo dataNode, sep string) string {
	names := make([]string, 0)

	// If these panic it's a programming error as we changed the document type while processing
	nativeNode := node.(*protobufquery.Node)
	nativeRelativeTo := relativeTo.(*protobufquery.Node)

	// Climb up the tree and collect the node names
	n := nativeNode.Parent
	for n != nil && n != nativeRelativeTo {
		kind := reflect.Invalid
		if n.Parent != nil && n.Parent.Value() != nil {
			kind = reflect.TypeOf(n.Parent.Value()).Kind()
		}
		switch kind {
		case reflect.Slice, reflect.Array:
			// Determine the index for array elements
			names = append(names, d.index(n))
		default:
			// Use the name if not an array
			names = append(names, n.Name)
		}
		n = n.Parent
	}

	if len(names) < 1 {
		return ""
	}

	// Construct the nodes
	nodepath := ""
	for _, name := range names {
		nodepath = name + sep + nodepath
	}

	return nodepath[:len(nodepath)-1]
}

func (d *protobufDocument) GetNodeName(node dataNode, sep string, withParent bool) string {
	// If this panics it's a programming error as we changed the document type while processing
	nativeNode := node.(*protobufquery.Node)

	name := nativeNode.Name

	// Check if the node is part of an array. If so, determine the index and
	// concatenate the parent name and the index.
	kind := reflect.Invalid
	if nativeNode.Parent != nil && nativeNode.Parent.Value() != nil {
		kind = reflect.TypeOf(nativeNode.Parent.Value()).Kind()
	}

	switch kind {
	case reflect.Slice, reflect.Array:
		if name == "" && nativeNode.Parent != nil && withParent {
			name = nativeNode.Parent.Name + sep
		}
		return name + d.index(nativeNode)
	}

	return name
}

func (*protobufDocument) OutputXML(node dataNode) string {
	native := node.(*protobufquery.Node)
	return native.OutputXML()
}

func (*protobufDocument) index(node *protobufquery.Node) string {
	idx := 0

	for n := node; n.PrevSibling != nil; n = n.PrevSibling {
		idx++
	}

	return strconv.Itoa(idx)
}
