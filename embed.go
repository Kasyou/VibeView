package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/renderer/*
var rendererFS embed.FS

func rendererAssets() http.Handler {
	sub, err := fs.Sub(rendererFS, "web/renderer")
	if err != nil {
		panic("vibeview: embedded renderer assets not found: " + err.Error())
	}
	return http.FileServer(http.FS(sub))
}

func rendererHTMLBytes() []byte {
	data, err := rendererFS.ReadFile("web/renderer/index.html")
	if err != nil {
		panic("vibeview: embedded renderer index.html not found: " + err.Error())
	}
	return data
}
