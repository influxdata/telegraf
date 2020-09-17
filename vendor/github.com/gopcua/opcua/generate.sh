#!/bin/sh

rm -f */*_gen.go
go run cmd/id/main.go
go run cmd/status/main.go
go run cmd/service/*.go

# install stringer if not installed already
command -v stringer || go get -u golang.org/x/tools/cmd/stringer

# find all enum types
enums=$(grep -w '^type' ua/enums*.go | awk '{print $2;}' | paste -sd, -)

# generate enum string method
(cd ua && stringer -type $enums -output enums_strings_gen.go)
echo "Wrote ua/enums_strings_gen.go"

# remove golang.org/x/tools/cmd/stringer from list of dependencies
go mod tidy
