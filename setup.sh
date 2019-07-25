#!/bin/sh

. common.inc.sh

run_cmd go get github.com/jteeuwen/go-bindata/...
run_cmd go get github.com/elazarl/go-bindata-assetfs/...
run_cmd go get github.com/gorilla/mux
run_cmd go get github.com/PuerkitoBio/goquery
