#!/bin/bash

gox -output="releases/{{.Dir}}_{{.OS}}_{{.Arch}}" -os="linux windows"
gox -output="releases/{{.Dir}}_{{.OS}}_{{.Arch}}" -osarch="darwin/amd64"