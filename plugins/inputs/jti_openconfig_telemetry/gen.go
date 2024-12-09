package jti_openconfig_telemetry

// To run these commands, make sure that protoc-gen-go and protoc-gen-go-grpc are installed
// > go install google.golang.org/protobuf/cmd/protoc-gen-go
// > go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
//
// Generated files were last generated with:
// - protoc-gen-go: v1.27.1
// - protoc-gen-go-grpc: v1.1.0
//go:generate protoc --go_out=auth/ --go-grpc_out=auth/ auth/authentication_service.proto
//go:generate protoc --go_out=oc/ --go-grpc_out=oc/ oc/oc.proto
