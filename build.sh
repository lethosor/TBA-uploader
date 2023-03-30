#!/usr/bin/env bash

cd "$(dirname "$0")"
. common.inc.sh

run_cmd node consts-gen.js consts.json --output-go consts.go --output-js web/src/consts.js
run_cmd node consts-gen.js consts.json --output-go tba/consts.go --go-package tba
test -n "$build_js" && run_cmd yarn run --silent build --stats errors-only
run_cmd cp README.md web/dist/
rm -vf bindata_assetfs.go  # silent migration
run_cmd go build -o bin/TBA-uploader
