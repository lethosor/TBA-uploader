#!/usr/bin/env bash

cd "$(dirname "$0")"
. common.inc.sh

run_cmd go get -u github.com/gorilla/mux
run_cmd go get -u github.com/PuerkitoBio/goquery
run_cmd go get -u github.com/go-test/deep
run_cmd yarn install
