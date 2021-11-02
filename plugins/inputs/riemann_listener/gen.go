package riemann_listener

// To run this command, make sure that protoc-gen-go is installed
// > go install google.golang.org/protobuf/cmd/protoc-gen-go
//
// Ensure that "go mod vendor" has been run before executing this generate command
//
// Generated file was last generated with:
// - protoc-gen-go: v1.27.1
//go:generate protoc --proto_path=../../../vendor/github.com/riemann/riemann-go-client --go_out=. --go_opt=Mproto/proto.proto=./riemangoProto proto/proto.proto
