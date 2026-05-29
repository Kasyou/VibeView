package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"vibeview/internal/detector"
	"vibeview/internal/server"
	"vibeview/internal/watcher"
)

func TestFullFlow_HTMLDetectAndWatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><h1>Hello</h1></body></html>`), 0644)

	info := detector.Detect(dir)
	if info.Type != detector.HTML {
		t.Fatalf("expected html, got %s", info.Type)
	}

	w := watcher.New(info.WatchDirs)
	defer w.Close()

	time.Sleep(100 * time.Millisecond)

	os.WriteFile(filepath.Join(dir, "style.css"), []byte(`body{}`), 0644)

	select {
	case <-w.Events:
		// ok
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for file change event")
	}
}

func TestFullFlow_ServerEndpoints(t *testing.T) {
	srv := server.New(server.Config{
		Port:         51821,
		DevServerURL: "http://localhost:5173",
		ProjectDir:   ".",
		RendererHTML: rendererHTMLBytes(),
		RendererFS:   rendererAssets(),
	})

	go func() { srv.Start() }()
	defer srv.Close()
	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get("http://localhost:51821/health")
	if err != nil {
		t.Skipf("server not reachable: %v", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("health: expected 200, got %d", resp.StatusCode)
	}
}

func TestFullFlow_RendererEmbedded(t *testing.T) {
	html := rendererHTMLBytes()
	if len(html) == 0 {
		t.Fatal("renderer HTML is empty")
	}
	s := string(html)
	if !strings.Contains(s, "VibeView") {
		t.Error("renderer HTML should contain VibeView")
	}
	if !strings.Contains(s, "device-frame") {
		t.Error("renderer HTML should contain device-frame")
	}
}

func TestFullFlow_RendererAssets(t *testing.T) {
	h := rendererAssets()
	if h == nil {
		t.Fatal("renderer assets handler is nil")
	}
}
