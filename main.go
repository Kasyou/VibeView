package main

import (
	"fmt"
	"os"
	"os/signal"

	"vibeview/internal/detector"
	"vibeview/internal/mcp"
	"vibeview/internal/server"
	"vibeview/internal/watcher"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		runMCP()
		return
	}

	projectDir, _ := os.Getwd()
	info := detector.Detect(projectDir)

	devURL := info.DevServerURL
	if info.ServeLocal {
		devURL = "http://localhost:51820/_app/index.html"
	}

	fmt.Println("  VibeView v0.1.0")
	fmt.Printf("  Project:   %s\n", info.Type)
	fmt.Printf("  Dev URL:   %s\n", devURL)

	srv := server.New(server.Config{
		Port:         51820,
		DevServerURL: devURL,
		ProjectDir:   projectDir,
		RendererHTML: rendererHTMLBytes(),
		RendererFS:   rendererAssets(),
	})

	w := watcher.New(info.WatchDirs)
	defer w.Close()

	go func() {
		for range w.Events {
			srv.Broadcast("reload", nil)
		}
	}()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		fmt.Println("\n  Shutting down...")
		os.Exit(0)
	}()

	fmt.Printf("  Preview:   http://localhost:%d\n", 51820)
	fmt.Printf("  Watching:  %v\n", info.WatchDirs)
	fmt.Println()

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runMCP() {
	serverURL := "http://localhost:51820"
	if envURL := os.Getenv("VIBEVIEW_URL"); envURL != "" {
		serverURL = envURL
	}

	mcpServer := mcp.New(serverURL)
	if err := mcpServer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP error: %v\n", err)
		os.Exit(1)
	}
}
