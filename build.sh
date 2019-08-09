#!/usr/bin/env bash

. common.inc.sh

cd $(dirname "$0")
test -n "$build_js" && run_cmd yarn run --silent build --display errors-only
run_cmd cp README.md web/dist/
run_cmd go-bindata -fs -prefix web/dist/ -o bindata_assetfs.go web/dist/...
run_cmd go build -o bin/TBA-uploader
