package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"vibeview/internal/detector"
	"vibeview/internal/mcp"
	"vibeview/internal/server"
	"vibeview/internal/term"
	"vibeview/internal/watcher"
)

const version = "0.1.0"

func main() {
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

	// Resolve project directory
	projectDir := *dir
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s %v\n", term.RedText("Error:"), err)
			os.Exit(1)
		}
	}
	projectDir, err := filepath.Abs(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s cannot resolve path: %v\n", term.RedText("Error:"), err)
		os.Exit(1)
	}

	// Validate directory exists
	if fi, err := os.Stat(projectDir); err != nil || !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "  %s directory not found: %s\n", term.RedText("Error:"), projectDir)
		os.Exit(1)
	}

	info := detector.Detect(projectDir)

	devURL := info.DevServerURL
	if info.ServeLocal {
		devURL = fmt.Sprintf("http://localhost:%d/_app/index.html", *port)
	}

	fmt.Println("  " + term.PinkText(term.BoldText("VibeView")) + " " + term.DimText("v"+version))
	fmt.Printf("  %s  %s\n", term.DimText("Project:"), term.CyanText(string(info.Type)))
	if info.ServeLocal {
		fmt.Printf("  %s   %s\n", term.DimText("Mode:"), term.GreenText("local serve"))
	} else {
		fmt.Printf("  %s    %s\n", term.DimText("Dev:"), term.CyanText(devURL))
	}

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
			if info.ServeLocal {
				srv.Broadcast("reload", nil)
			}
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Println("\n  " + term.DimText("Shutting down..."))
		w.Close()
		srv.Close()
		os.Exit(0)
	}()

	fmt.Printf("  %s  %s\n", term.DimText("Preview:"), term.GreenText(fmt.Sprintf("http://localhost:%d", *port)))
	fmt.Printf("  %s %v\n", term.DimText("Watching:"), info.WatchDirs)
	fmt.Println()

	if err := srv.Start(); err != nil {
		// Friendly error for port already in use
		if strings.Contains(err.Error(), "bind") ||
			strings.Contains(err.Error(), "address already in use") ||
			strings.Contains(err.Error(), "in use") {
			fmt.Fprintf(os.Stderr, "  %s port %d is already in use.\n", term.RedText("Error:"), *port)
			fmt.Fprintf(os.Stderr, "  %s\n", term.DimText("Try: vibeview --port " + fmt.Sprintf("%d", *port+1)))
		} else {
			fmt.Fprintf(os.Stderr, "  %s %v\n", term.RedText("Error:"), err)
		}
		os.Exit(1)
	}
}

func runMCP() {
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
