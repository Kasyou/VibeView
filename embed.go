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

func rendererHTMLBytes(mode string) []byte {
	file := "web/renderer/index.html"
	if mode == "claude" {
		file = "web/renderer/claude.html"
	}
	data, err := rendererFS.ReadFile(file)
	if err != nil {
		panic("vibeview: embedded renderer not found: " + file + ": " + err.Error())
	}
	return data
}
