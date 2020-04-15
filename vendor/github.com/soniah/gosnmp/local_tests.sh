#!/bin/bash

go test -v -tags helper
go test -v -tags marshal
go test -v -tags misc
go test -v -tags api
go test -v -tags trap
