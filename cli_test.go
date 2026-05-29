package main

import (
	"bytes"
	"flag"
	"os"
	"testing"
)

func TestHelpOutput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHelp()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if out == "" {
		t.Error("help output should not be empty")
	}
	if !contains(out, "VibeView") {
		t.Error("help should contain VibeView")
	}
	if !contains(out, "--port") {
		t.Error("help should document --port flag")
	}
	if !contains(out, "--dir") {
		t.Error("help should document --dir flag")
	}
	if !contains(out, "mcp") {
		t.Error("help should mention mcp command")
	}
}

func TestFlagParsingDefaults(t *testing.T) {
	// Reset flags between tests
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	port := fs.Int("port", 51820, "")
	dir := fs.String("dir", "", "")

	fs.Parse([]string{})

	if *port != 51820 {
		t.Errorf("default port should be 51820, got %d", *port)
	}
	if *dir != "" {
		t.Errorf("default dir should be empty, got %s", *dir)
	}
}

func TestFlagParsingCustom(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	port := fs.Int("port", 51820, "")
	dir := fs.String("dir", "", "")

	fs.Parse([]string{"--port", "3000", "--dir", "/tmp/test"})

	if *port != 3000 {
		t.Errorf("port should be 3000, got %d", *port)
	}
	if *dir != "/tmp/test" {
		t.Errorf("dir should be /tmp/test, got %s", *dir)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
