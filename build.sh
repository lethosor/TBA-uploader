#!/bin/sh
set -e

run_cmd() {
    echo "=> $@"
    "$@"
}

cd $(dirname "$0")
run_cmd go build -o bin/TBA-uploader
