#!/bin/sh

. common.inc.sh

run_cmd go get -u github.com/go-bindata/go-bindata/...
run_cmd go get -u github.com/gorilla/mux
run_cmd go get -u github.com/PuerkitoBio/goquery
run_cmd yarn install
