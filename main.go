package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		runMCP()
		return
	}
	fmt.Println("VibeView v0.1.0")
	fmt.Println("Preview:  http://localhost:51820")
}

func runMCP() {
	fmt.Fprintln(os.Stderr, "MCP mode (not yet implemented)")
}
