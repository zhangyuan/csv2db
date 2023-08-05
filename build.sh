#!/bin/bash

set -euo pipefail

BINARY_NAME=${BINARY_NAME:-csv2sqlite3}

rm -rf bin/${BINARY_NAME}*

CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "-w" -o bin/${BINARY_NAME}-amd64.exe

CGO_ENABLED=1 GOOS=windows GOARCH=386 go build -ldflags "-w" -o bin/${BINARY_NAME}-386.exe

CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "-w" -o bin/${BINARY_NAME}-amd64-darwin

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-w" -o bin/${BINARY_NAME}-amd64-linux

CGO_ENABLED=1 GOOS=linux GOARCH=386 go build -ldflags "-w" -o bin/${BINARY_NAME}-386-linux
