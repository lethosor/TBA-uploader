package main

import (
	"embed"
	"io/fs"
)

const ASSETS_BASE_PATH = "tmp/assets/"

//go:generate rm -rf tmp/assets/
//go:generate mkdir -p tmp/assets
//go:generate cp -r ../web/build tmp/assets/web
//go:embed all:tmp/assets/web/*
var fsWebEmbedded embed.FS

var fsWeb fs.FS

func init() {
	var err error
	webPath := ASSETS_BASE_PATH + "web"
	fsWeb, err = fs.Sub(fsWebEmbedded, webPath)
	if err != nil {
		panic("missing assets at path: " + webPath)
	}
}
