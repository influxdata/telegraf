package xpath

import (
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/antchfx/xpath"
	"github.com/doclambda/protobufquery"

	// Register all known definitions APIv2 types
	// "github.com/doclambda/protobufquery/testcases/addressbook"

	// Register all known definitions APIv1 types
	_ "github.com/doclambda/protobufquery/testcases/addressbook"
)

var once sync.Once

type protobufDocument struct {
	MessageType string
	msg         *dynamicpb.Message
}

func (d *protobufDocument) Init() error {
	var err error

	// Register all APIv2 style packages. APIv1 style packages will register themselves, so nothing to do.
	once.Do(func() {
		// if err = protoregistry.GlobalFiles.RegisterFile(addressbook.File_addressbook_proto); err != nil { return }
	})
	if err != nil {
		return fmt.Errorf("registering APIv2 protocol-buffer failed: %v", err)
	}

	// Check the message type
	if d.MessageType == "" {
		return fmt.Errorf("protocol-buffer message-type not set")
	}

	// Get a prototypical message for later use
	msgFullName := protoreflect.FullName(d.MessageType)
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(msgFullName)
	if err != nil {
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

func (d *protobufDocument) CreateXPathNavigator(node dataNode) xpath.NodeNavigator {
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
	path := ""
	for _, name := range names {
		path = name + sep + path
	}

	return path[:len(path)-1]
}

func (d *protobufDocument) OutputXML(node dataNode) string {
	native := node.(*protobufquery.Node)
	return native.OutputXML()
}

func init() {
}
