#!/bin/sh

run_cmd() {
    echo "=> $@"
    "$@"
}

cd $(dirname "$0")
run_cmd go build -o bin/TBA-uploader
