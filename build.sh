#!/bin/sh

. common.inc.sh

cd $(dirname "$0")
run_cmd cp README.md web/
run_cmd go-bindata-assetfs web/... -o bindata_assetfs.go
run_cmd go build -o bin/TBA-uploader
