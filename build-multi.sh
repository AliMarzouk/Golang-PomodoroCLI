#!/bin/bash
set -e

APP=pomodorocli

# Linux
GOOS=linux GOARCH=amd64 go build -o dist/$APP-linux-amd64 .

# Windows
GOOS=windows GOARCH=amd64 go build -o dist/$APP-windows-amd64.exe .

# macOS
GOOS=darwin GOARCH=arm64 go build -o dist/$APP-macos-arm64 .
GOOS=darwin GOARCH=amd64 go build -o dist/$APP-macos-amd64 .
