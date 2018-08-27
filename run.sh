#!/bin/sh

run_cmd() {
    echo "=> $@"
    "$@"
}

cd $(dirname "$0")
./build.sh
run_cmd ./bin/TBA-uploader "$@"
