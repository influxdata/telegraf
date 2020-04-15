proto.pb.go: proto.proto
	mkdir -p _pb
	protoc --go_out=_pb $<
	cat _pb/$@\
	|gofmt >$@
	rm -rf _pb
