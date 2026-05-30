package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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
		case "design":
			runPreviewMode("design")
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

func runPreview()          { runPreviewMode("claude") }
func portFilePath(projectDir string) string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".vibeview")
	os.MkdirAll(dir, 0755)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(projectDir)))
	return filepath.Join(dir, hash)
}
func runPreviewMode(defaultMode string) {
	fs := flag.NewFlagSet("vibeview", flag.ExitOnError)
	port := fs.Int("port", 0, "HTTP server port (default: 51820 Claude, 51821 Design)")
	mode := fs.String("mode", defaultMode, "Mode: claude (AI whiteboard) or design (instant preview)")
	ontop := fs.Bool("ontop", false, "Print PowerShell command to pin preview window always-on-top")
	dir := fs.String("dir", "", "Project directory (default: current directory)")

	// Skip first arg if it's a subcommand (design, mcp, etc.)
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "design" {
		args = args[1:]
	}
	fs.Parse(args)

	if *port == 0 {
		if *mode == "design" {
			*port = 51821
		} else {
			*port = 51820
		}
	}

	label := "Claude"
	if *mode == "design" {
		label = "Design"
	}

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

	fmt.Println("  " + term.PinkText(term.BoldText("VibeView")) + " " + term.DimText("v"+version) + "  " + term.CyanText("["+label+"]"))
	fmt.Printf("  %s  %s\n", term.DimText("Project:"), term.CyanText(string(info.Type)))
	if *mode == "design" {
		fmt.Printf("  %s   %s\n", term.DimText("Mode:"), term.GreenText("instant preview"))
	} else if info.ServeLocal {
		fmt.Printf("  %s   %s\n", term.DimText("Mode:"), term.GreenText("local serve + MCP"))
	} else {
		fmt.Printf("  %s    %s (MCP)\n", term.DimText("Dev:"), term.CyanText(devURL))
	}

	srv := server.New(server.Config{
		Port:         *port,
		DevServerURL: devURL,
		ProjectDir:   projectDir,
		Mode:         *mode,
		RendererHTML: rendererHTMLBytes(*mode),
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

	// Auto-retry next port if busy, print URL only after success
	for attempt := 0; attempt < 10; attempt++ {
		os.WriteFile(portFilePath(projectDir), []byte(fmt.Sprintf("%d", *port)), 0644)
		err := srv.Start()
		if err == nil {
			fmt.Printf("  %s  %s\n", term.DimText("Preview:"), term.GreenText(fmt.Sprintf("http://localhost:%d", *port)))
			fmt.Printf("  %s %v\n", term.DimText("Watching:"), info.WatchDirs)
			if *ontop {
				fmt.Println()
				fmt.Println("  " + term.BoldText("To pin window always-on-top:"))
				fmt.Println("    " + term.DimText("PowerToys:") + " Win+Ctrl+T on the browser window")
			}
			fmt.Println()
			break
		}
		if attempt < 9 && (strings.Contains(err.Error(), "bind") ||
			strings.Contains(err.Error(), "address already in use") ||
			strings.Contains(err.Error(), "in use")) {
			fmt.Printf("  %s port %d busy, trying %d...\n", term.DimText("→"), *port, *port+1)
			*port++
			devURL = fmt.Sprintf("http://localhost:%d/_app/index.html", *port)
			srv = server.New(server.Config{
				Port: *port, DevServerURL: devURL, ProjectDir: projectDir,
				Mode: *mode, RendererHTML: rendererHTMLBytes(*mode), RendererFS: rendererAssets(),
			})
			continue
		}
		if strings.Contains(err.Error(), "bind") ||
			strings.Contains(err.Error(), "address already in use") ||
			strings.Contains(err.Error(), "in use") {
			fmt.Fprintf(os.Stderr, "  %s all ports %d-%d are in use.\n", term.RedText("Error:"), *port-9, *port)
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
	// Read port from centralized file ($HOME/.vibeview/<hash>)
	if cwd, err := os.Getwd(); err == nil {
		if data, err := os.ReadFile(portFilePath(cwd)); err == nil {
			if p, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && p > 0 {
				serverURL = fmt.Sprintf("http://localhost:%d", p)
			}
		}
	}

	mcpServer := mcp.New(serverURL)
	if err := mcpServer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "MCP error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`VibeView v` + version + ` — instant UI preview & Claude visual whiteboard.

Usage:
  vibeview [--mode claude|design] [--port N] [--dir PATH] [--ontop]
  vibeview mcp [--port N]
  vibeview help
  vibeview version

Modes:
  claude (default)  Claude whiteboard — visualize AI reasoning as styled cards
  design            Instant preview — see UI changes as you code

Flags:
  --mode    Mode: claude (AI collaboration) or design (instant preview)
  --port    Server port (default: 51820 claude, 51821 design)
  --dir     Project directory (default: current dir)
  --ontop   Print PowerShell command to pin window always-on-top

Examples:
  vibeview                              Claude whiteboard on :51820
  vibeview --mode design                Design preview on :51821
  vibeview --mode design --ontop        Design preview + always-on-top help
  vibeview --port 3000 --dir ./my-app   Custom port and project
  vibeview mcp                          MCP server for Claude Code`)
}
