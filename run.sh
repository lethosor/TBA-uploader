#!/bin/sh
set -e

run_cmd() {
    echo "=> $@"
    "$@"
}

cd $(dirname "$0")
test -n "$build" && ./build.sh
run_cmd ./bin/TBA-uploader "$@" -data-folder ./fms_data -web-folder ./web
