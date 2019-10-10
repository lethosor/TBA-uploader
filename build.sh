#!/usr/bin/env bash

cd "$(dirname "$0")"
. common.inc.sh

run_cmd node consts-gen.js consts.json --output-go consts.go --output-js web/src/consts.js
test -n "$build_js" && run_cmd yarn run --silent build --display errors-only
run_cmd cp README.md web/dist/
run_cmd go-bindata -fs -prefix web/dist/ -o bindata_assetfs.go web/dist/...
run_cmd go build -o bin/TBA-uploader
