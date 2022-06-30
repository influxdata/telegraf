package xpath

import (
	"fmt"
	"sort"
	"strings"

	"github.com/influxdata/telegraf"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"

	path "github.com/antchfx/xpath"
	"github.com/doclambda/protobufquery"
)

type protobufDocument struct {
	MessageDefinition string
	MessageType       string
	ImportPaths       []string
	Log               telegraf.Logger
	msg               *dynamicpb.Message
}

func (d *protobufDocument) Init() error {
	// Check the message definition and type
	if d.MessageDefinition == "" {
		return fmt.Errorf("protocol-buffer message-definition not set")
	}
	if d.MessageType == "" {
		return fmt.Errorf("protocol-buffer message-type not set")
	}

	// Load the file descriptors from the given protocol-buffer definition
	parser := protoparse.Parser{
		ImportPaths:      d.ImportPaths,
		InferImportPaths: true,
	}
	fds, err := parser.ParseFiles(d.MessageDefinition)
	if err != nil {
		return fmt.Errorf("parsing protocol-buffer definition in %q failed: %v", d.MessageDefinition, err)
	}
	if len(fds) < 1 {
		return fmt.Errorf("file %q does not contain file descriptors", d.MessageDefinition)
	}

	// Register all definitions in the file in the global registry
	registry, err := protodesc.NewFiles(desc.ToFileDescriptorSet(fds...))
	if err != nil {
		return fmt.Errorf("constructing registry failed: %v", err)
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
	if err := proto.Unmarshal(buf, msg.Interface()); err != nil {
		return nil, err
	}

	return protobufquery.Parse(msg)
}

func (d *protobufDocument) QueryAll(node dataNode, expr string) ([]dataNode, error) {
	// If this panics it's a programming error as we changed the document type while processing
	native, err := protobufquery.QueryAll(node.(*protobufquery.Node), expr)
	if err != nil {
		return nil, err
	}

	nodes := make([]dataNode, len(native))
	for i, n := range native {
		nodes[i] = n
	}
	return nodes, nil
}

func (d *protobufDocument) CreateXPathNavigator(node dataNode) path.NodeNavigator {
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
		names = append(names, n.Name)
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

func (d *protobufDocument) OutputXML(node dataNode) string {
	native := node.(*protobufquery.Node)
	return native.OutputXML()
}

func init() {
}
