package xpath

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"

	path "github.com/antchfx/xpath"
	"github.com/doclambda/protobufquery"

	// Register all known definitions
	_ "github.com/doclambda/protobufquery/testcases/addressbook"
	_ "github.com/doclambda/tahu/client_libraries/golang"
)

var once sync.Once

type protobufDocument struct {
	MessageType string
	Log         telegraf.Logger
	msg         *dynamicpb.Message
}

func (d *protobufDocument) Init() error {
	var err error

	// Register all packages requiring manual registration. Usually packages register themselves on import.
	once.Do(func() {
		// if err = protoregistry.GlobalFiles.RegisterFile(tahu.File_sparkplug_b_proto); err != nil { return }
	})
	if err != nil {
		return fmt.Errorf("registering protocol-buffers manually failed: %v", err)
	}

	// Check the message type
	if d.MessageType == "" {
		return fmt.Errorf("protocol-buffer message-type not set")
	}

	// Get a prototypical message for later use
	msgFullName := protoreflect.FullName(d.MessageType)
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(msgFullName)
	if err != nil {
		d.Log.Infof("Could not find %q... Known messages:", msgFullName)
		var known []string
		protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
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

	msgDesc, ok := desc.(protoreflect.MessageDescriptor)
	if !ok {
		return fmt.Errorf("found type for %q is not a message descriptor (%T)", msgFullName, desc)
	}

	d.msg = dynamicpb.NewMessage(msgDesc)
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
