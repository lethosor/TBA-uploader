#!/usr/bin/env bash

cd "$(dirname "$0")"
. ../common.inc.sh

run_cmd go build -o ../bin/ minimize-test-html.go
