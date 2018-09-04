#!/bin/sh

GOFMT_FILES=$(gofmt -l .)
if [ -n "${GOFMT_FILES}" ]; then
  printf >&2 'gofmt failed for the following files:\n%s\n\nplease run "gofmt -w ." on your changes before committing.\n' "${GOFMT_FILES}"
  exit 1
fi

GOLINT_ERRORS=$(golint ./... | grep -v "Id should be")
if [ -n "${GOLINT_ERRORS}" ]; then
  printf >&2 'golint failed for the following reasons:\n%s\n\nplease run 'golint ./...' on your changes before committing.\n' "${GOLINT_ERRORS}"
  exit 1
fi

GOVET_ERRORS=$(go tool vet *.go 2>&1)
if [ -n "${GOVET_ERRORS}" ]; then
  printf >&2 'go vet failed for the following reasons:\n%s\n\nplease run "go tool vet *.go" on your changes before committing.\n' "${GOVET_ERRORS}"
  exit 1
fi

if [ -z "${NOTEST}" ]; then
  printf >&2 'Running short tests...\n'
  env AMQP_URL= go test -short -v | egrep 'PASS|ok'

  if [ $? -ne 0 ]; then
    printf >&2 'go test failed, please fix before committing.\n'
    exit 1
  fi
fi
