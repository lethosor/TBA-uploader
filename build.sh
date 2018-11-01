#!/bin/sh
set -e

run_cmd() {
    echo "=> $@"
    "$@"
}

cd $(dirname "$0")
run_cmd cp README.md web/
run_cmd go-bindata-assetfs web/...
run_cmd go build -i -o bin/TBA-uploader
