package cisco_telemetry_mdt

// To run these commands, make sure that protoc-gen-go and protoc-gen-go-grpc are installed
// > go install google.golang.org/protobuf/cmd/protoc-gen-go
// > go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
//
// Generated files were last generated with:
// - protoc-gen-go: v1.27.1
// - protoc-gen-go-grpc: v1.1.0
//go:generate protoc --go_out=dialout/ --go-grpc_out=dialout/ dialout/dialout.proto
//go:generate protoc --go_out=telemetry/ telemetry/telemetry.proto
