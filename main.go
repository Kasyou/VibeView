package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"vibeview/internal/detector"
	"vibeview/internal/mcp"
	"vibeview/internal/server"
	"vibeview/internal/watcher"
)

const version = "0.1.0"

func main() {
	// Handle subcommands / help
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "mcp":
			runMCP()
			return
		case "help", "-h", "--help":
			printHelp()
			return
		case "version", "-v", "--version":
			fmt.Println("VibeView v" + version)
			return
		}
	}

	runPreview()
}

func runPreview() {
	port := flag.Int("port", 51820, "HTTP server port")
	dir := flag.String("dir", "", "Project directory (default: current directory)")
	flag.CommandLine.Parse(os.Args[1:])

	projectDir := *dir
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
	projectDir, _ = filepath.Abs(projectDir)

	info := detector.Detect(projectDir)

	devURL := info.DevServerURL
	if info.ServeLocal {
		devURL = fmt.Sprintf("http://localhost:%d/_app/index.html", *port)
	}

	fmt.Println("  VibeView v" + version)
	fmt.Printf("  Project:   %s\n", info.Type)
	fmt.Printf("  Dev URL:   %s\n", devURL)

	srv := server.New(server.Config{
		Port:         *port,
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
		srv.Close()
		os.Exit(0)
	}()

	fmt.Printf("  Preview:   http://localhost:%d\n", *port)
	fmt.Printf("  Watching:  %v\n", info.WatchDirs)
	fmt.Println()

	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runMCP() {
	// parse flags from os.Args[2:] (skip "mcp")
	port := flag.Int("port", 51820, "Preview server port")
	flag.CommandLine.Parse(os.Args[2:])

	serverURL := fmt.Sprintf("http://localhost:%d", *port)
	if envURL := os.Getenv("VIBEVIEW_URL"); envURL != "" {
		serverURL = envURL
	}

	mcpServer := mcp.New(serverURL)
	if err := mcpServer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`VibeView v` + version + ` — instant UI preview for vibe coding.

Usage:
  vibeview [--port N] [--dir PATH]     Start preview server
  vibeview mcp [--port N]              Start MCP server (for Claude Code)
  vibeview help                        Show this help
  vibeview version                     Show version

Flags:
  --port int    Server port (default: 51820)
  --dir  path   Project directory (default: current dir)

Examples:
  vibeview                        Start in current directory
  vibeview --port 3000            Start on port 3000
  vibeview --dir ~/my-project     Start for a specific project
  vibeview mcp                    Start MCP server for Claude Code`)
}
