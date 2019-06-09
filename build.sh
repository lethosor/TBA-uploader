#!/bin/sh
set -e

if [ "$(uname)" = "Darwin" ]; then
    # need to use system compiler
    unset CC
    unset CXX
fi

run_cmd() {
    echo "=> $@"
    "$@"
}

cd $(dirname "$0")
run_cmd cp README.md web/
run_cmd go-bindata-assetfs web/...
run_cmd go build -o bin/TBA-uploader
