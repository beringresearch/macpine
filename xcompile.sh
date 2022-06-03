#!/bin/bash
set -e
# you may need to go install github.com/mitchellh/gox@v1.0.1 first

CGO_ENABLED=1 gox -ldflags "${LDFLAGS}" -output="bin/alpine_{{.OS}}_{{.Arch}}" --osarch="darwin/amd64 darwin/arm64" 
