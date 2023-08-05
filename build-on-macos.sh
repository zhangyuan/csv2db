#!/bin/bash

set -euo pipefail

BINARY_NAME=${BINARY_NAME:-csv2sqlite3}

rm -rf bin/${BINARY_NAME}*

GOOS=darwin GOARCH=amd64 go build -ldflags "-w" -o bin/${BINARY_NAME}-amd64-darwin
