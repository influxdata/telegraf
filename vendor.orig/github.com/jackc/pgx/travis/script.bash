#!/usr/bin/env bash
set -eux

if [ "${PGVERSION-}" != "" ]
then
  go test -v -race ./...
elif [ "${CRATEVERSION-}" != "" ]
then
  go test -v -race -run 'TestCrateDBConnect'
fi
