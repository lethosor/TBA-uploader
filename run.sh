#!/usr/bin/env bash

. common.inc.sh

cd $(dirname "$0")
test -n "$build" && ./build.sh
run_cmd ./bin/TBA-uploader "$@" -data-folder ./fms_data -web-folder ./web/dist
