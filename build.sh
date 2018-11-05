#!/bin/bash

gox -ldflags="-s -w" -output="release/{{.Dir}}_{{.OS}}_{{.Arch}}" -osarch="darwin/amd64 linux/amd64 linux/386 linux/arm64 linux/arm windows/386 windows/amd64"