#!/usr/bin/env bash

cd "$(dirname "$0")"
. common.inc.sh

test -n "$build" && ./build.sh
run_cmd ./bin/TBA-uploader "$@" -data-folder ./fms_data -web-folder ./web/dist
