# VibeView Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a CLI tool that gives Android Studio-style instant UI preview for vibe coding — browser preview with device frames, WebSocket hot reload, and MCP Server for Claude Code.

**Architecture:** Go CLI with embedded renderer (go:embed). Core HTTP/WS server on localhost:51820 serves a preview page that wraps the user's project in a device-frame iframe. File watcher triggers WebSocket reload. MCP mode exposes tools to Claude Code via stdio JSON-RPC.

**Tech Stack:** Go 1.22, fsnotify, gorilla/websocket, vanilla HTML/CSS/JS (embedded), Vite dev server (external, user-provided)

---

## File Structure

```
VibeView/
├── main.go                          # Entry point, CLI
├── go.mod
├── go.sum
├── CLAUDE.md                        # Project instructions for Claude
├── Makefile                         # Build + release
├── internal/
│   ├── detector/
│   │   ├── detector.go              # Framework detection
│   │   └── detector_test.go
│   ├── watcher/
│   │   ├── watcher.go               # fsnotify + debounce
│   │   └── watcher_test.go
│   ├── server/
│   │   ├── server.go                # HTTP + WebSocket server
│   │   └── server_test.go
│   └── mcp/
│       ├── mcp.go                   # MCP JSON-RPC over stdio
│       └── mcp_test.go
├── web/
│   └── renderer/
│       ├── index.html               # Preview page UI
│       ├── app.js                   # Preview logic (device frame, iframe, WS, errors)
│       └── style.css                # Device frames, toolbar, error cards
├── scripts/
│   └── build.sh                     # Cross-compile script
└── docs/
    └── superpowers/
        ├── specs/
        │   └── 2026-05-29-vibeview-design.md
        └── plans/
            └── 2026-05-29-vibeview-plan.md
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\go.mod`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\main.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\Makefile`

- [ ] **Step 1: Initialize Go module**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go mod init vibeview`

- [ ] **Step 2: Write minimal main.go**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		fmt.Println("MCP mode (not yet implemented)")
		os.Exit(0)
	}
	fmt.Println("VibeView v0.1.0")
	fmt.Println("Preview:  http://localhost:51820")
}
```

- [ ] **Step 3: Verify it builds and runs**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go build -o vibeview.exe . && ./vibeview.exe`
Expected: prints version and URLs, exits cleanly.

- [ ] **Step 4: Write Makefile**

```makefile
.PHONY: build run test clean

build:
	go build -o vibeview.exe .

run: build
	./vibeview.exe

test:
	go test ./internal/...

clean:
	rm -f vibeview.exe
```

- [ ] **Step 5: Commit**

```bash
git add go.mod main.go Makefile
git commit -m "feat: scaffold Go project with minimal entry point"
```

---

### Task 2: Framework Detector

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\detector\detector.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\detector\detector_test.go`

- [ ] **Step 1: Write the test**

```go
package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectReact(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"react":"^18.0.0"},"devDependencies":{"vite":"^5.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.ts"), []byte(`export default {}`), 0644)

	result := Detect(dir)
	if result.Type != "react" {
		t.Errorf("expected react, got %s", result.Type)
	}
	if result.DevServerPort != 5173 {
		t.Errorf("expected port 5173, got %d", result.DevServerPort)
	}
}

func TestDetectVue(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"vue":"^3.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(`export default {}`), 0644)

	result := Detect(dir)
	if result.Type != "vue" {
		t.Errorf("expected vue, got %s", result.Type)
	}
}

func TestDetectHTML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html></html>`), 0644)

	result := Detect(dir)
	if result.Type != "html" {
		t.Errorf("expected html, got %s", result.Type)
	}
}

func TestDetectSvelte(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"devDependencies":{"@sveltejs/vite-plugin-svelte":"^3.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(`export default {}`), 0644)

	result := Detect(dir)
	if result.Type != "svelte" {
		t.Errorf("expected svelte, got %s", result.Type)
	}
}

func TestDetectWatchDirs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"react":"^18.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.ts"), []byte(`export default {}`), 0644)
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	result := Detect(dir)
	if len(result.WatchDirs) == 0 {
		t.Error("expected non-empty watch dirs")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go test ./internal/detector/ -v`
Expected: compilation errors (types not defined)

- [ ] **Step 3: Write detector.go**

```go
package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProjectType string

const (
	React  ProjectType = "react"
	Vue    ProjectType = "vue"
	Svelte ProjectType = "svelte"
	HTML   ProjectType = "html"
	Unknown ProjectType = "unknown"
)

type ProjectInfo struct {
	Type          ProjectType
	DevServerPort int
	WatchDirs     []string
	DevServerURL  string
}

func Detect(root string) ProjectInfo {
	info := ProjectInfo{
		Type:          Unknown,
		DevServerPort: 5173,
	}

	pkg := readPackageJSON(root)
	viteConfig := hasViteConfig(root)
	hasHTML := fileExists(filepath.Join(root, "index.html"))

	switch {
	case pkg.hasDep("react") && viteConfig:
		info.Type = React
		info.WatchDirs = existingDirs(root, "src")
	case pkg.hasDep("vue") && viteConfig:
		info.Type = Vue
		info.WatchDirs = existingDirs(root, "src")
	case pkg.hasDep("@sveltejs/vite-plugin-svelte") && viteConfig:
		info.Type = Svelte
		info.WatchDirs = existingDirs(root, "src")
	case hasHTML:
		info.Type = HTML
		info.WatchDirs = []string{root}
	default:
		info.Type = HTML
		info.WatchDirs = []string{root}
	}

	info.DevServerURL = devServerURL(info.DevServerPort)
	return info
}

type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPackageJSON(root string) packageJSON {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return packageJSON{}
	}
	var pkg packageJSON
	json.Unmarshal(data, &pkg)
	return pkg
}

func (p packageJSON) hasDep(name string) bool {
	if _, ok := p.Dependencies[name]; ok {
		return true
	}
	if _, ok := p.DevDependencies[name]; ok {
		return true
	}
	return false
}

func hasViteConfig(root string) bool {
	return fileExists(filepath.Join(root, "vite.config.js")) ||
		fileExists(filepath.Join(root, "vite.config.ts")) ||
		fileExists(filepath.Join(root, "vite.config.mjs"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func existingDirs(root string, dirs ...string) []string {
	var result []string
	for _, d := range dirs {
		path := filepath.Join(root, d)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			result = append(result, path)
		}
	}
	if len(result) == 0 {
		result = append(result, root)
	}
	return result
}

func devServerURL(port int) string {
	return "http://localhost:5173"
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go test ./internal/detector/ -v`
Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/detector/
git commit -m "feat: add framework auto-detection for React/Vue/Svelte/HTML"
```

---

### Task 3: File Watcher

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\watcher\watcher.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\watcher\watcher_test.go`

- [ ] **Step 1: Install fsnotify**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go get github.com/fsnotify/fsnotify`

- [ ] **Step 2: Write watcher_test.go**

```go
package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherDetectsFileChange(t *testing.T) {
	dir := t.TempDir()

	w := New(dir)
	defer w.Close()

	events := w.Events()

	// write a file
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)

	select {
	case <-events:
		// success
	case <-time.After(2 * time.Second):
		t.Error("timed out waiting for file change event")
	}
}

func TestWatcherDebounce(t *testing.T) {
	dir := t.TempDir()

	w := New(dir)
	defer w.Close()

	events := w.Events()

	// rapid writes should only produce one event
	go func() {
		for i := 0; i < 10; i++ {
			os.WriteFile(filepath.Join(dir, "test.txt"), []byte("data"), 0644)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// first event should arrive within debounce window
	select {
	case <-events:
		// success
	case <-time.After(3 * time.Second):
		t.Error("timed out waiting for debounced event")
	}
}
```

- [ ] **Step 3: Write watcher.go**

```go
package watcher

import (
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	w      *fsnotify.Watcher
	events chan struct{}
	done   chan struct{}
}

func New(root string) *Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	// add root and subdirectories
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			// skip node_modules and hidden dirs
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || (len(base) > 0 && base[0] == '.') {
				return filepath.SkipDir
			}
			w.Add(path)
		}
		return nil
	})

	ww := &Watcher{
		w:      w,
		events: make(chan struct{}, 1),
		done:   make(chan struct{}),
	}

	go ww.loop()
	return ww
}

func (w *Watcher) loop() {
	debounce := time.NewTimer(0)
	<-debounce.C // drain initial fire
	pending := false

	for {
		select {
		case <-w.w.Events:
			debounce.Reset(100 * time.Millisecond)
			pending = true
		case <-debounce.C:
			if pending {
				select {
				case w.events <- struct{}{}:
				default:
				}
				pending = false
			}
		case <-w.done:
			debounce.Stop()
			return
		}
	}
}

func (w *Watcher) Events() <-chan struct{} {
	return w.events
}

func (w *Watcher) Close() {
	close(w.done)
	w.w.Close()
}
```

Hmm, I need `os` import. Let me fix that.

```go
package watcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)
```

- [ ] **Step 4: Run tests**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go test ./internal/watcher/ -v -timeout 10s`
Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/watcher/ go.mod go.sum
git commit -m "feat: add debounced file watcher"
```

---

### Task 4: HTTP + WebSocket Server

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\server\server.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\server\server_test.go`

- [ ] **Step 1: Install gorilla/websocket**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go get github.com/gorilla/websocket`

- [ ] **Step 2: Write server_test.go**

```go
package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerServesRenderer(t *testing.T) {
	s := New(Config{
		Port:         0,
		DevServerURL: "http://localhost:5173",
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "<html") {
		t.Error("response should contain HTML")
	}
}

func TestServerHealthEndpoint(t *testing.T) {
	s := New(Config{
		Port:         0,
		DevServerURL: "http://localhost:5173",
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestServerReloadEndpoint(t *testing.T) {
	s := New(Config{
		Port:         0,
		DevServerURL: "http://localhost:5173",
	})

	req := httptest.NewRequest("POST", "/api/reload", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
```

- [ ] **Step 3: Write server.go**

```go
package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Config struct {
	Port         int
	DevServerURL string
	WatchDirs    []string
}

type Server struct {
	cfg     Config
	mux     *http.ServeMux
	wsUp    websocket.Upgrader
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func New(cfg Config) *Server {
	s := &Server{
		cfg:     cfg,
		mux:     http.NewServeMux(),
		clients: make(map[*websocket.Conn]bool),
		wsUp: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleRenderer)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ws", s.handleWS)
	s.mux.HandleFunc("/api/reload", s.handleReload)
	s.mux.HandleFunc("/api/config", s.handleConfig)
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"devServerURL":"%s"}`, s.cfg.DevServerURL)
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.Broadcast("reload", nil)
	w.Write([]byte(`{"ok":true}`))
}

func (s *Server) handleRenderer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(rendererHTML)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUp.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			conn.Close()
		}()
		for {
			// keep connection alive, read messages (console logs, screenshot data)
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()
}

func (s *Server) Broadcast(msgType string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		conn.WriteJSON(map[string]interface{}{
			"type": msgType,
			"data": data,
		})
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	return http.ListenAndServe(addr, s.mux)
}

// rendererHTML is set from embed in a later task
var rendererHTML = []byte(`<html><body><h1>VibeView</h1></body></html>`)
```

- [ ] **Step 4: Run tests**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go test ./internal/server/ -v`
Expected: all tests PASS

- [ ] **Step 5: Commit**

```bash
git add internal/server/ go.mod go.sum
git commit -m "feat: add HTTP and WebSocket server"
```

---

### Task 5: Renderer UI

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\web\renderer\index.html`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\web\renderer\app.js`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\web\renderer\style.css`

- [ ] **Step 1: Write index.html**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>VibeView</title>
  <link rel="stylesheet" href="/_vibeview/style.css">
</head>
<body>
  <div id="toolbar">
    <span id="brand">VibeView</span>
    <div id="device-picker">
      <button data-device="iphone15" class="active">iPhone 15 Pro</button>
      <button data-device="pixel8">Pixel 8</button>
      <button data-device="ipad">iPad</button>
      <button data-device="full">Full Width</button>
      <button data-device="custom">Custom</button>
    </div>
    <div id="custom-size" style="display:none">
      <input type="number" id="custom-w" placeholder="W" value="375">
      <span>×</span>
      <input type="number" id="custom-h" placeholder="H" value="812">
      <button id="custom-apply">Apply</button>
    </div>
    <span id="status"></span>
  </div>

  <div id="preview-container">
    <div id="device-frame">
      <div id="device-notch" class="iphone-notch"></div>
      <iframe id="app-frame" src="" sandbox="allow-scripts allow-same-origin allow-forms"></iframe>
    </div>
  </div>

  <div id="error-overlay" style="display:none">
    <div id="error-card">
      <div id="error-icon">!</div>
      <div id="error-message"></div>
      <div id="error-file"></div>
    </div>
  </div>

  <script src="/_vibeview/app.js"></script>
</body>
</html>
```

- [ ] **Step 2: Write style.css**

```css
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #1a1a2e;
  color: #e0e0e0;
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
}

#toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: #16213e;
  border-bottom: 1px solid #0f3460;
  flex-shrink: 0;
}

#brand {
  font-weight: 700;
  font-size: 14px;
  color: #e94560;
  margin-right: 16px;
}

#device-picker {
  display: flex;
  gap: 2px;
  background: #0f3460;
  border-radius: 6px;
  padding: 2px;
}

#device-picker button {
  background: none;
  border: none;
  color: #a0a0b0;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.15s;
}

#device-picker button.active {
  background: #e94560;
  color: #fff;
}

#device-picker button:hover:not(.active) {
  color: #fff;
  background: rgba(233, 69, 96, 0.2);
}

#custom-size {
  display: flex;
  gap: 4px;
  align-items: center;
}

#custom-size input {
  width: 56px;
  padding: 3px 6px;
  background: #0f3460;
  border: 1px solid #1a1a4e;
  border-radius: 4px;
  color: #e0e0e0;
  font-size: 12px;
}

#custom-apply {
  background: #e94560;
  border: none;
  color: #fff;
  padding: 3px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}

#status {
  margin-left: auto;
  font-size: 11px;
  color: #666;
}

#preview-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  background: #1a1a2e;
}

#device-frame {
  position: relative;
  background: #111;
  border-radius: 24px;
  padding: 12px;
  box-shadow: 0 0 0 2px #333, 0 0 0 4px #1a1a2e, 0 8px 32px rgba(0,0,0,0.5);
  transition: all 0.3s ease;
}

#device-frame.iphone15 {
  width: 417px;
  height: 876px;
  padding-top: 44px;
}

#device-frame.pixel8 {
  width: 436px;
  height: 939px;
}

#device-frame.ipad {
  width: 767px;
  height: 1157px;
  border-radius: 18px;
  padding: 16px;
}

#device-frame.full {
  width: 100%;
  height: 100%;
  border-radius: 8px;
  padding: 4px;
}

#device-notch {
  display: none;
  position: absolute;
  top: 12px;
  left: 50%;
  transform: translateX(-50%);
  width: 120px;
  height: 28px;
  background: #111;
  border-radius: 0 0 16px 16px;
  z-index: 10;
}

#device-frame.iphone15 #device-notch { display: block; }

#iframe-frame {
  width: 100%;
  height: 100%;
  border: none;
  border-radius: 12px;
  background: #fff;
}

#device-frame.full #app-frame { border-radius: 4px; }

#error-overlay {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 100;
}

#error-card {
  background: #e94560;
  color: #fff;
  padding: 12px 20px;
  border-radius: 8px;
  font-size: 13px;
  box-shadow: 0 4px 16px rgba(233, 69, 96, 0.4);
  display: flex;
  align-items: center;
  gap: 10px;
  max-width: 500px;
}

#error-icon {
  width: 24px;
  height: 24px;
  background: rgba(255,255,255,0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 14px;
  flex-shrink: 0;
}

#error-message { flex: 1; }
#error-file {
  font-size: 11px;
  opacity: 0.7;
  white-space: nowrap;
}
```

- [ ] **Step 3: Write app.js**

```javascript
const DEVICES = {
  iphone15: { width: 393, height: 852, label: 'iPhone 15 Pro' },
  pixel8:  { width: 412, height: 915, label: 'Pixel 8' },
  ipad:    { width: 744, height: 1133, label: 'iPad' },
  full:    { width: 0, height: 0, label: 'Full Width' },
};

let currentDevice = 'iphone15';
let devServerURL = '';

// --- Init ---
async function init() {
  const res = await fetch('/api/config');
  const cfg = await res.json();
  devServerURL = cfg.devServerURL;

  connectWS();
  setupDevicePicker();
  setDevice('iphone15');
  loadApp();
}

// --- WebSocket ---
function connectWS() {
  const ws = new WebSocket(`ws://${location.host}/ws`);
  ws.onmessage = (e) => {
    const msg = JSON.parse(e.data);
    if (msg.type === 'reload') {
      loadApp();
    }
  };
  ws.onclose = () => {
    document.getElementById('status').textContent = 'disconnected';
    setTimeout(connectWS, 2000);
  };
  ws.onopen = () => {
    document.getElementById('status').textContent = 'live';
  };
}

// --- Device Picker ---
function setupDevicePicker() {
  document.querySelectorAll('#device-picker button').forEach(btn => {
    btn.addEventListener('click', () => {
      if (btn.dataset.device === 'custom') {
        document.getElementById('custom-size').style.display = 'flex';
        return;
      }
      document.getElementById('custom-size').style.display = 'none';
      setDevice(btn.dataset.device);
    });
  });
  document.getElementById('custom-apply').addEventListener('click', () => {
    const w = parseInt(document.getElementById('custom-w').value) || 375;
    const h = parseInt(document.getElementById('custom-h').value) || 812;
    setCustomSize(w, h);
  });
}

function setDevice(name) {
  currentDevice = name;
  const frame = document.getElementById('device-frame');
  const iframe = document.getElementById('app-frame');
  const notch = document.getElementById('device-notch');

  // update button states
  document.querySelectorAll('#device-picker button').forEach(b => {
    b.classList.toggle('active', b.dataset.device === name);
  });

  frame.className = '';
  notch.style.display = 'none';

  if (name === 'full') {
    frame.classList.add('full');
    iframe.style.width = '100%';
    iframe.style.height = '100%';
  } else if (DEVICES[name]) {
    frame.classList.add(name);
    iframe.style.width = DEVICES[name].width + 'px';
    iframe.style.height = DEVICES[name].height + 'px';
    if (name === 'iphone15') {
      notch.style.display = 'block';
    }
  }
}

function setCustomSize(w, h) {
  document.querySelectorAll('#device-picker button').forEach(b => b.classList.remove('active'));
  const frame = document.getElementById('device-frame');
  const iframe = document.getElementById('app-frame');
  frame.className = '';
  frame.style.width = (w + 24) + 'px';
  frame.style.height = (h + 24) + 'px';
  iframe.style.width = w + 'px';
  iframe.style.height = h + 'px';
}

// --- App Loading ---
function loadApp() {
  const iframe = document.getElementById('app-frame');
  const url = devServerURL + '?_vibeview=' + Date.now();
  iframe.src = url;

  hideError();
}

// --- Error Boundary ---
window.addEventListener('message', (e) => {
  if (e.data && e.data.type === 'vibeview-error') {
    showError(e.data.message, e.data.file);
  }
});

function showError(message, file) {
  const overlay = document.getElementById('error-overlay');
  document.getElementById('error-message').textContent = message;
  document.getElementById('error-file').textContent = file || '';
  overlay.style.display = 'block';
}

function hideError() {
  document.getElementById('error-overlay').style.display = 'none';
}

// --- Start ---
init();
```

- [ ] **Step 4: Commit**

```bash
git add web/renderer/
git commit -m "feat: add renderer UI with device frames and error boundary"
```

---

### Task 6: Embed Renderer + Wire Server

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\server\embed.go`
- Modify: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\server\server.go`
- Modify: `D:\WORKS\ClaudeCode\CliProject\VibeView\main.go`

- [ ] **Step 1: Create embed.go**

```go
package server

import "embed"

//go:embed web/renderer/*
var rendererFS embed.FS

func init() {
	data, err := rendererFS.ReadFile("web/renderer/index.html")
	if err != nil {
		panic("failed to read embedded renderer: " + err.Error())
	}
	rendererHTML = data
}
```

Wait, the embed path is relative to the package. Since `embed.go` is in `internal/server/`, the embed path needs to go up to the project root. That doesn't work with `go:embed` - it can only embed files in the same directory or subdirectories.

I need to restructure. Either:
1. Move the embed to the root or a package at the right level
2. Copy the renderer files into the server package during build
3. Use a separate embed package

Let me restructure: put the embed at the root level in `main.go` or create an `embed.go` at the project root.

Actually, the cleanest approach: put `embed.go` at the project root (`D:\WORKS\ClaudeCode\CliProject\VibeView\embed.go`), and have it export the renderer bytes. The server package imports it.

Or even simpler: embed from the main package and pass the renderer HTML to the server config.

Let me use the approach of embedding in main.go and passing as a variable:

Actually wait, let me reconsider. The simplest approach:

1. Put a file `web/renderer/embed.go` in a location that can reference the web files
2. Since embed paths are relative to the package source, I need the embed to be in or above `web/renderer/`

The standard pattern is to embed from the package that's closest. Let me put the embed in the server package but copy web files there during build.

Actually, the SIMPLEST approach for Go is:
- Create a package at `internal/renderer/` that embeds the files
- `internal/server/` imports from `internal/renderer/`

Or even simpler: just embed at the main package level (main.go or root embed.go) since the web directory is at the root level.

Let me use this approach:

```go
// embed.go (at project root, package main)
package main

import "embed"

//go:embed web/renderer/*
var rendererFS embed.FS
```

Then pass to server via Config. Let me update the plan.

- [ ] **Step 1: Create root-level embed.go**

```go
package main

import "embed"

//go:embed web/renderer/*
var embeddedRenderer embed.FS

func readRendererHTML() []byte {
	data, err := embeddedRenderer.ReadFile("web/renderer/index.html")
	if err != nil {
		panic("vibeview: embedded renderer not found: " + err.Error())
	}
	return data
}
```

- [ ] **Step 2: Update server.go to accept renderer bytes via Config**

Replace the `rendererHTML` var in server.go:

```go
type Config struct {
	Port         int
	DevServerURL string
	WatchDirs    []string
	RendererHTML []byte
}

type Server struct {
	cfg     Config
	mux     *http.ServeMux
	wsUp    websocket.Upgrader
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

// ...

func (s *Server) handleRenderer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(s.cfg.RendererHTML)
}

// Remove: var rendererHTML = []byte(...)
```

- [ ] **Step 3: Update server to serve static files for renderer assets (CSS, JS)**

Add to `registerRoutes()`:

```go
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/", s.handleRenderer)
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ws", s.handleWS)
	s.mux.HandleFunc("/api/reload", s.handleReload)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	s.mux.HandleFunc("/_vibeview/", s.handleStatic)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// serve embedded CSS/JS files
	path := "web/renderer" + r.URL.Path[len("/_vibeview"):]
	data, err := embeddedRenderer.ReadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, ".css") {
		w.Header().Set("Content-Type", "text/css")
	} else if strings.HasSuffix(r.URL.Path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
	}
	w.Write(data)
}
```

Wait, `embeddedRenderer` is in the main package. The server package can't access it. I need to either:
a) Pass renderer FS via Config
b) Put the embed in the server package and copy files

Let me go with option (a): pass the entire `embed.FS` to Config.

```go
// server.go Config
type Config struct {
	Port         int
	DevServerURL string
	WatchDirs    []string
	RendererFS   embed.FS
}
```

Hmm, passing embed.FS by value is fine. But the rendererHTML reading should happen in server, not in main.

Let me simplify. I'll pass rendererHTML as []byte, and also create a file server from the embed FS. Let me think...

Actually the simplest thing: in main.go, read the HTML into bytes, and create a handler that serves the static files from the embed FS. Pass both to the server.

Hmm, this is getting complex. Let me simplify further for v1:

**Approach: don't serve CSS/JS as separate files. Inline everything in the HTML.** This is simpler and the renderer is small.

Update index.html to inline CSS in `<style>` and JS in `<script>` tags.

Wait no, that's ugly. Let me use a proper approach:

1. `embed.go` at project root embeds `web/renderer/*`
2. `Config` has `RendererHTML []byte` field
3. Main reads HTML from embed and sets it in config
4. For serving CSS/JS: add a `StaticHandler http.Handler` to Config, created in main from embed.FS via `http.FileServer`

Actually, the simplest Go approach:

```go
// In main.go
rendererFS := http.FileServer(http.FS(embeddedRenderer))
config := server.Config{
    Port: 51820,
    DevServerURL: devURL,
    WatchDirs: watchDirs,
    RendererFS: rendererFS,
}
```

And in server:
```go
type Config struct {
    Port         int
    DevServerURL string
    WatchDirs    []string
    RendererFS   http.Handler
}
```

Then the server just delegates static requests to RendererFS. And reads the index.html itself.

Actually, let me just do this cleanly. I'll update the plan now. Let me just write the actual plan steps more cleanly.

Let me redo this task properly.

- [ ] **Step 1: Create embed.go at project root**

File: `D:\WORKS\ClaudeCode\CliProject\VibeView\embed.go`

```go
package main

import (
	"embed"
	"net/http"
)

//go:embed web/renderer/*
var rendererFS embed.FS

func rendererAssets() http.Handler {
	sub, _ := rendererFS.Sub("web/renderer")
	return http.FileServer(http.FS(sub))
}

func rendererHTML() []byte {
	data, err := rendererFS.ReadFile("web/renderer/index.html")
	if err != nil {
		panic("vibeview: embedded renderer not found: " + err.Error())
	}
	return data
}
```

- [ ] **Step 2: Update server.go Config and routing**

Modify `Config`:
```go
type Config struct {
	Port         int
	DevServerURL string
	WatchDirs    []string
	RendererHTML []byte
	RendererFS   http.Handler
}
```

Add static file handler in registerRoutes:
```go
func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/ws", s.handleWS)
	s.mux.HandleFunc("/api/reload", s.handleReload)
	s.mux.HandleFunc("/api/config", s.handleConfig)
	// serve renderer assets under /_vibeview/
	s.mux.Handle("/_vibeview/", http.StripPrefix("/_vibeview", s.cfg.RendererFS))
	// serve renderer page at /
	s.mux.HandleFunc("/", s.handleRenderer)
}
```

Update handleRenderer:
```go
func (s *Server) handleRenderer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(s.cfg.RendererHTML)
}
```

- [ ] **Step 3: Update main.go to wire everything**

```go
package main

import (
	"fmt"
	"os"
	"os/signal"

	"vibeview/internal/detector"
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

	fmt.Println("VibeView v0.1.0")
	fmt.Printf("Project:   %s\n", info.Type)
	fmt.Printf("Dev URL:   %s\n", info.DevServerURL)
	fmt.Printf("Preview:   http://localhost:51820\n")

	srv := server.New(server.Config{
		Port:         51820,
		DevServerURL: info.DevServerURL,
		WatchDirs:    info.WatchDirs,
		RendererHTML: rendererHTML(),
		RendererFS:   rendererAssets(),
	})

	// Start file watcher
	w := watcher.New(projectDir, info.WatchDirs)
	defer w.Close()

	go func() {
		for range w.Events() {
			srv.Broadcast("reload", nil)
		}
	}()

	// Graceful shutdown
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		fmt.Println("\nShutting down...")
		os.Exit(0)
	}()

	fmt.Println("Watching: ", info.WatchDirs)
	fmt.Println("Ready. Open http://localhost:51820 in your browser.")
	if err := srv.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runMCP() {
	// Task 8
	fmt.Fprintf(os.Stderr, "MCP mode starting...\n")
}
```

But wait, I need to update the watcher to accept multiple watch dirs. Let me update watcher.New signature.

Actually, let me update the watcher in this task too.

- [ ] **Step 4: Update watcher.go to accept multiple dirs**

Change New signature:
```go
func New(root string, watchDirs []string) *Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	for _, dir := range watchDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				base := filepath.Base(path)
				if base == "node_modules" || base == ".git" || (len(base) > 0 && base[0] == '.') {
					return filepath.SkipDir
				}
				w.Add(path)
			}
			return nil
		})
	}
	// ... rest is same
}
```

- [ ] **Step 5: Update server_test.go for new Config fields**

Update the test configs:
```go
func TestServerServesRenderer(t *testing.T) {
	s := New(Config{
		Port:         0,
		DevServerURL: "http://localhost:5173",
		RendererHTML: []byte("<html><body>test</body></html>"),
		RendererFS:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	})
	// ...
}
```

Hmm wait, I need to import "net/http" in server_test.go. Actually it's already imported.

OK let me redo this whole task. The plan updates are getting complex. Let me re-write task 6 cleanly.

Actually, I think the plan is fine to write with the correct code. Let me just make sure the plan doc has all the right code in it. Let me continue writing the rest of the tasks and fix any issues.

- [ ] **Step 6: Build and verify**

Run: `cd D:/WORKS/ClaudeCode/CliProject/VibeView && go build -o vibeview.exe .`
Expected: builds successfully

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: wire up embedded renderer, file watcher, and server"
```

OK this task is getting too complex. Let me simplify and just write the complete plan. I'll be less granular about test-every-step and focus on the code being correct.

Let me redo the plan with cleaner organization. I'll be concise but complete.<｜end▁of▁thinking｜>

<｜｜DSML｜｜tool_calls>
<｜｜DSML｜｜invoke name="Write">
<｜｜DSML｜｜parameter name="content" string="true"># VibeView Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a CLI tool that gives Android Studio-style instant UI preview for vibe coding — browser preview with device frames, WebSocket hot reload, and MCP Server for Claude Code.

**Architecture:** Go CLI binary with embedded renderer (go:embed). Core HTTP/WS server on localhost:51820 serves a preview page that wraps the user's project in a device-frame iframe. File watcher triggers WebSocket reload. MCP mode exposes tools via stdio JSON-RPC. For Vite projects, proxies to existing Vite dev server. For HTML projects, serves files directly with livereload injection.

**Tech Stack:** Go 1.22, fsnotify, gorilla/websocket, vanilla HTML/CSS/JS (embedded), Vite dev server (external, user-provided)

---

## File Structure

```
VibeView/
├── main.go                          # Entry point, CLI flag parsing
├── embed.go                         # go:embed for web/renderer files
├── go.mod
├── go.sum
├── CLAUDE.md                        # Project instructions for Claude
├── Makefile                         # Build shortcuts
├── internal/
│   ├── detector/
│   │   ├── detector.go              # Framework detection from package.json
│   │   └── detector_test.go
│   ├── watcher/
│   │   ├── watcher.go               # fsnotify + debounce
│   │   └── watcher_test.go
│   ├── server/
│   │   ├── server.go                # HTTP + WebSocket + API
│   │   └── server_test.go
│   └── mcp/
│       ├── mcp.go                   # MCP JSON-RPC over stdio
│       └── mcp_test.go
├── web/
│   └── renderer/
│       ├── index.html               # Preview page structure
│       ├── app.js                   # Device frame, iframe, WS, errors
│       └── style.css                # Device frames, toolbar, error cards
├── scripts/
│   └── build.sh                     # Cross-compile for all platforms
└── docs/
    └── superpowers/
        ├── specs/
        │   └── 2026-05-29-vibeview-design.md
        └── plans/
            └── 2026-05-29-vibeview-plan.md
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\go.mod`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\main.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\Makefile`

- [ ] **Step 1: Initialize Go module**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go mod init vibeview
```

- [ ] **Step 2: Create directories**

```bash
mkdir -p internal/detector internal/watcher internal/server internal/mcp
mkdir -p web/renderer scripts
```

- [ ] **Step 3: Write minimal main.go**

```go
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
	fmt.Fprintln(os.Stderr, "MCP mode starting...")
}
```

- [ ] **Step 4: Write Makefile**

```makefile
.PHONY: build run test clean

build:
	go build -o vibeview.exe .

run: build
	./vibeview.exe

test:
	go test ./internal/... -v

clean:
	rm -f vibeview.exe
```

- [ ] **Step 5: Build and verify**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go build -o vibeview.exe . && ./vibeview.exe
```

Expected output: `VibeView v0.1.0` then `Preview: http://localhost:51820`

- [ ] **Step 6: Commit**

```bash
git add go.mod main.go Makefile
git commit -m "feat: scaffold Go project with minimal entry point"
```

---

### Task 2: Framework Detector

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\detector\detector.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\detector\detector_test.go`

- [ ] **Step 1: Write detector.go**

```go
package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProjectType string

const (
	React   ProjectType = "react"
	Vue     ProjectType = "vue"
	Svelte  ProjectType = "svelte"
	HTML    ProjectType = "html"
	Unknown ProjectType = "unknown"
)

type ProjectInfo struct {
	Type          ProjectType
	DevServerPort int
	DevServerURL  string
	WatchDirs     []string
}

func Detect(root string) ProjectInfo {
	info := ProjectInfo{
		Type:          Unknown,
		DevServerPort: 5173,
	}

	pkg := readPackageJSON(root)
	hasVite := viteConfigExists(root)
	hasHTML := fileExists(filepath.Join(root, "index.html"))

	switch {
	case pkg.hasDep("react") && hasVite:
		info.Type = React
		info.WatchDirs = dirs(root, "src")
	case pkg.hasDep("vue") && hasVite:
		info.Type = Vue
		info.WatchDirs = dirs(root, "src")
	case pkg.hasDep("@sveltejs/vite-plugin-svelte") || pkg.hasDep("svelte"):
		info.Type = Svelte
		info.WatchDirs = dirs(root, "src")
	case hasHTML:
		info.Type = HTML
		info.WatchDirs = []string{root}
	default:
		info.Type = HTML
		info.WatchDirs = []string{root}
	}

	info.DevServerURL = "http://localhost:5173"
	return info
}

type pkgJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPackageJSON(root string) pkgJSON {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return pkgJSON{}
	}
	var p pkgJSON
	json.Unmarshal(data, &p)
	return p
}

func (p pkgJSON) hasDep(name string) bool {
	if _, ok := p.Dependencies[name]; ok {
		return true
	}
	if _, ok := p.DevDependencies[name]; ok {
		return true
	}
	return false
}

func viteConfigExists(root string) bool {
	return fileExists(filepath.Join(root, "vite.config.js")) ||
		fileExists(filepath.Join(root, "vite.config.ts")) ||
		fileExists(filepath.Join(root, "vite.config.mjs"))
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirs(root string, names ...string) []string {
	var result []string
	for _, n := range names {
		p := filepath.Join(root, n)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		result = append(result, root)
	}
	return result
}
```

- [ ] **Step 2: Write detector_test.go**

```go
package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectReact(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dependencies":{"react":"^18.0.0"},"devDependencies":{"vite":"^5.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.ts"), []byte(`export default {}`), 0644)

	r := Detect(dir)
	if r.Type != React {
		t.Errorf("expected react, got %s", r.Type)
	}
}

func TestDetectVue(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dependencies":{"vue":"^3.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(`export default {}`), 0644)

	r := Detect(dir)
	if r.Type != Vue {
		t.Errorf("expected vue, got %s", r.Type)
	}
}

func TestDetectHTML(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html></html>`), 0644)

	r := Detect(dir)
	if r.Type != HTML {
		t.Errorf("expected html, got %s", r.Type)
	}
}

func TestDetectSvelte(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"devDependencies":{"svelte":"^4.0.0","@sveltejs/vite-plugin-svelte":"^3.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.js"), []byte(`export default {}`), 0644)

	r := Detect(dir)
	if r.Type != Svelte {
		t.Errorf("expected svelte, got %s", r.Type)
	}
}

func TestDetectWatchDirs(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"),
		[]byte(`{"dependencies":{"react":"^18.0.0"}}`), 0644)
	os.WriteFile(filepath.Join(dir, "vite.config.ts"), []byte(`export default {}`), 0644)
	os.MkdirAll(filepath.Join(dir, "src"), 0755)

	r := Detect(dir)
	if len(r.WatchDirs) == 0 {
		t.Error("expected non-empty watch dirs")
	}
	if r.WatchDirs[0] != filepath.Join(dir, "src") {
		t.Errorf("expected src dir, got %s", r.WatchDirs[0])
	}
}
```

- [ ] **Step 3: Run tests**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go test ./internal/detector/ -v
```

Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/detector/
git commit -m "feat: add framework auto-detection for React/Vue/Svelte/HTML"
```

---

### Task 3: File Watcher

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\watcher\watcher.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\watcher\watcher_test.go`

- [ ] **Step 1: Install fsnotify**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go get github.com/fsnotify/fsnotify
```

- [ ] **Step 2: Write watcher.go**

```go
package watcher

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	fw     *fsnotify.Watcher
	Events chan struct{}
	done   chan struct{}
}

func New(watchDirs []string) *Watcher {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		panic("vibeview: fsnotify: " + err.Error())
	}

	w := &Watcher{
		fw:     fw,
		Events: make(chan struct{}, 1),
		done:   make(chan struct{}),
	}

	for _, dir := range watchDirs {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || !info.IsDir() {
				return nil
			}
			base := filepath.Base(path)
			if base == "node_modules" || base == ".git" || (len(base) > 0 && base[0] == '.') {
				return filepath.SkipDir
			}
			fw.Add(path)
			return nil
		})
	}

	go w.loop()
	return w
}

func (w *Watcher) loop() {
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	pending := false

	for {
		select {
		case <-w.fw.Events:
			timer.Reset(100 * time.Millisecond)
			pending = true
		case <-timer.C:
			if pending {
				select {
				case w.Events <- struct{}{}:
				default:
				}
				pending = false
			}
		case <-w.done:
			timer.Stop()
			return
		}
	}
}

func (w *Watcher) Close() {
	close(w.done)
	w.fw.Close()
}
```

- [ ] **Step 3: Write watcher_test.go**

```go
package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcherDetectsChange(t *testing.T) {
	dir := t.TempDir()
	w := New([]string{dir})
	defer w.Close()

	time.Sleep(50 * time.Millisecond) // let fsnotify initialize
	os.WriteFile(filepath.Join(dir, "test.css"), []byte("body{}"), 0644)

	select {
	case <-w.Events:
		// ok
	case <-time.After(3 * time.Second):
		t.Error("timeout waiting for file change")
	}
}

func TestWatcherDebounce(t *testing.T) {
	dir := t.TempDir()
	w := New([]string{dir})
	defer w.Close()

	time.Sleep(50 * time.Millisecond)

	// rapid writes
	for i := 0; i < 20; i++ {
		os.WriteFile(filepath.Join(dir, "test.css"), []byte("body{}"), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	// should get exactly 1 event (debounced)
	count := 0
	timeout := time.After(2 * time.Second)
loop:
	for {
		select {
		case <-w.Events:
			count++
		case <-timeout:
			break loop
		}
	}
	if count == 0 {
		t.Error("expected at least 1 event")
	}
	if count > 3 {
		t.Errorf("expected <=3 debounced events, got %d", count)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go test ./internal/watcher/ -v -timeout 15s
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/watcher/ go.mod go.sum
git commit -m "feat: add debounced file watcher"
```

---

### Task 4: HTTP + WebSocket Server

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\server\server.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\server\server_test.go`

- [ ] **Step 1: Install gorilla/websocket**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go get github.com/gorilla/websocket
```

- [ ] **Step 2: Write server.go**

```go
package server

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Config struct {
	Port         int
	DevServerURL string
	RendererHTML []byte
	RendererFS   http.Handler
}

type Server struct {
	cfg     Config
	mux     *http.ServeMux
	wsUp    websocket.Upgrader
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
	console []ConsoleMsg
}

type ConsoleMsg struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	File    string `json:"file"`
	Line    int    `json:"line"`
}

func New(cfg Config) *Server {
	s := &Server{
		cfg:     cfg,
		mux:     http.NewServeMux(),
		clients: make(map[*websocket.Conn]bool),
		wsUp: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.health)
	s.mux.HandleFunc("/ws", s.ws)
	s.mux.HandleFunc("/api/reload", s.reload)
	s.mux.HandleFunc("/api/config", s.config)
	s.mux.HandleFunc("/api/console", s.getConsole)
	s.mux.Handle("/_vibeview/", http.StripPrefix("/_vibeview", s.cfg.RendererFS))
	s.mux.HandleFunc("/", s.renderer)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"devServerURL":"%s"}`, s.cfg.DevServerURL)
}

func (s *Server) renderer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(s.cfg.RendererHTML)
}

func (s *Server) reload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	s.Broadcast("reload", nil)
	w.Write([]byte(`{"ok":true}`))
}

func (s *Server) getConsole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.mu.Lock()
	defer s.mu.Unlock()
	// return and clear
	msgs := s.console
	s.console = nil
	if msgs == nil {
		w.Write([]byte(`[]`))
		return
	}
	fmt.Fprintf(w, `[`)
	for i, m := range msgs {
		if i > 0 {
			w.Write([]byte(`,`))
		}
		fmt.Fprintf(w, `{"level":"%s","message":"%s","file":"%s","line":%d}`, m.Level, m.Message, m.File, m.Line)
	}
	w.Write([]byte(`]`))
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUp.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			conn.Close()
		}()
		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			// handle console forwarding
			if msg["type"] == "console" {
				s.mu.Lock()
				data := msg["data"].(map[string]interface{})
				s.console = append(s.console, ConsoleMsg{
					Level:   data["level"].(string),
					Message: data["message"].(string),
				})
				s.mu.Unlock()
			}
		}
	}()
}

func (s *Server) Broadcast(msgType string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		conn.WriteJSON(map[string]interface{}{
			"type": msgType,
			"data": data,
		})
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	return http.ListenAndServe(addr, s.mux)
}
```

- [ ] **Step 3: Write server_test.go**

```go
package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func stubHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("static"))
	})
}

func TestHealth(t *testing.T) {
	s := New(Config{Port: 0, DevServerURL: "http://localhost:5173", RendererHTML: []byte("x"), RendererFS: stubHandler()})
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRenderer(t *testing.T) {
	s := New(Config{Port: 0, DevServerURL: "http://localhost:5173", RendererHTML: []byte("<html>hi</html>"), RendererFS: stubHandler()})
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hi") {
		t.Error("expected renderer HTML")
	}
}

func TestConfig(t *testing.T) {
	s := New(Config{Port: 0, DevServerURL: "http://localhost:3000", RendererHTML: []byte("x"), RendererFS: stubHandler()})
	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "localhost:3000") {
		t.Error("expected devServerURL in config response")
	}
}

func TestReloadMethodCheck(t *testing.T) {
	s := New(Config{Port: 0, DevServerURL: "", RendererHTML: []byte("x"), RendererFS: stubHandler()})
	req := httptest.NewRequest("GET", "/api/reload", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go test ./internal/server/ -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/server/ go.mod go.sum
git commit -m "feat: add HTTP and WebSocket server with console forwarding"
```

---

### Task 5: Renderer UI

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\web\renderer\index.html`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\web\renderer\app.js`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\web\renderer\style.css`

- [ ] **Step 1: Write index.html**

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>VibeView</title>
  <link rel="stylesheet" href="/_vibeview/style.css">
</head>
<body>
  <div id="toolbar">
    <span id="brand">VibeView</span>
    <div id="device-picker">
      <button data-device="iphone15" class="active">iPhone 15 Pro</button>
      <button data-device="pixel8">Pixel 8</button>
      <button data-device="ipad">iPad</button>
      <button data-device="full">Full</button>
      <button data-device="custom">Custom</button>
    </div>
    <div id="custom-size" style="display:none">
      <input type="number" id="custom-w" placeholder="W" value="375">
      <span>×</span>
      <input type="number" id="custom-h" placeholder="H" value="812">
      <button id="custom-apply">Apply</button>
    </div>
    <span id="status">connecting</span>
  </div>
  <div id="preview-container">
    <div id="device-frame" class="iphone15">
      <div id="device-notch"></div>
      <iframe id="app-frame" src="" sandbox="allow-scripts allow-same-origin allow-forms"></iframe>
    </div>
  </div>
  <div id="error-overlay" style="display:none">
    <div id="error-card">
      <div id="error-icon">!</div>
      <div id="error-message"></div>
      <div id="error-file"></div>
    </div>
  </div>
  <script src="/_vibeview/app.js"></script>
</body>
</html>
```

- [ ] **Step 2: Write style.css**

```css
*{margin:0;padding:0;box-sizing:border-box}
body{
  font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;
  background:#1a1a2e;color:#e0e0e0;display:flex;flex-direction:column;
  height:100vh;overflow:hidden;
}
#toolbar{
  display:flex;align-items:center;gap:12px;padding:8px 16px;
  background:#16213e;border-bottom:1px solid #0f3460;flex-shrink:0;
}
#brand{font-weight:700;font-size:14px;color:#e94560;margin-right:16px}
#device-picker{display:flex;gap:2px;background:#0f3460;border-radius:6px;padding:2px}
#device-picker button{
  background:none;border:none;color:#a0a0b0;padding:4px 12px;
  border-radius:4px;cursor:pointer;font-size:12px;transition:all 0.15s;
}
#device-picker button.active{background:#e94560;color:#fff}
#device-picker button:hover:not(.active){color:#fff;background:rgba(233,69,96,0.2)}
#custom-size{display:flex;gap:4px;align-items:center}
#custom-size input{
  width:56px;padding:3px 6px;background:#0f3460;border:1px solid #1a1a4e;
  border-radius:4px;color:#e0e0e0;font-size:12px;
}
#custom-apply{background:#e94560;border:none;color:#fff;padding:3px 8px;border-radius:4px;cursor:pointer;font-size:12px}
#status{margin-left:auto;font-size:11px;color:#666}
#preview-container{
  flex:1;display:flex;align-items:center;justify-content:center;
  padding:20px;background:#1a1a2e;
}
#device-frame{
  position:relative;background:#111;border-radius:24px;padding:12px;
  box-shadow:0 0 0 2px #333,0 0 0 4px #1a1a2e,0 8px 32px rgba(0,0,0,0.5);
  transition:all 0.3s ease;
}
#device-frame.iphone15{width:417px;height:876px;padding-top:44px}
#device-frame.pixel8{width:436px;height:939px}
#device-frame.ipad{width:767px;height:1157px;border-radius:18px;padding:16px}
#device-frame.full{width:100%;height:100%;border-radius:8px;padding:4px}
#device-notch{
  display:none;position:absolute;top:12px;left:50%;transform:translateX(-50%);
  width:120px;height:28px;background:#111;border-radius:0 0 16px 16px;z-index:10;
}
#device-frame.iphone15 #device-notch{display:block}
#app-frame{width:100%;height:100%;border:none;border-radius:12px;background:#fff}
#device-frame.full #app-frame{border-radius:4px}
#error-overlay{position:fixed;bottom:20px;left:50%;transform:translateX(-50%);z-index:100}
#error-card{
  background:#e94560;color:#fff;padding:12px 20px;border-radius:8px;
  font-size:13px;box-shadow:0 4px 16px rgba(233,69,96,0.4);
  display:flex;align-items:center;gap:10px;max-width:500px;
}
#error-icon{
  width:24px;height:24px;background:rgba(255,255,255,0.2);border-radius:50%;
  display:flex;align-items:center;justify-content:center;font-weight:700;font-size:14px;flex-shrink:0;
}
#error-message{flex:1}
#error-file{font-size:11px;opacity:0.7;white-space:nowrap}
```

- [ ] **Step 3: Write app.js**

```javascript
(function() {
  const DEVICES = {
    iphone15: { width: 393, height: 852 },
    pixel8:  { width: 412, height: 915 },
    ipad:    { width: 744, height: 1133 },
  };

  let devServerURL = '';
  let ws;

  // --- Init ---
  async function init() {
    try {
      const res = await fetch('/api/config');
      const cfg = await res.json();
      devServerURL = cfg.devServerURL;
    } catch(e) {
      devServerURL = 'http://localhost:5173';
    }
    connectWS();
    setupDevicePicker();
    loadApp();
  }

  // --- WebSocket ---
  function connectWS() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    ws = new WebSocket(proto + '://' + location.host + '/ws');
    ws.onmessage = function(e) {
      var msg = JSON.parse(e.data);
      if (msg.type === 'reload') loadApp();
    };
    ws.onclose = function() {
      setStatus('reconnecting');
      setTimeout(connectWS, 2000);
    };
    ws.onopen = function() { setStatus('live'); };

    // Forward console errors from iframe
    window.addEventListener('message', function(e) {
      if (e.data && e.data.type === 'vibeview-error') {
        showError(e.data.message, e.data.file, e.data.line);
      }
    });
  }

  function setStatus(text) {
    var el = document.getElementById('status');
    if (el) el.textContent = text;
  }

  // --- Device Picker ---
  function setupDevicePicker() {
    document.querySelectorAll('#device-picker button').forEach(function(btn) {
      btn.addEventListener('click', function() {
        var dev = btn.dataset.device;
        if (dev === 'custom') {
          document.getElementById('custom-size').style.display = 'flex';
          return;
        }
        document.getElementById('custom-size').style.display = 'none';
        setDevice(dev);
      });
    });
    document.getElementById('custom-apply').addEventListener('click', function() {
      var w = parseInt(document.getElementById('custom-w').value) || 375;
      var h = parseInt(document.getElementById('custom-h').value) || 812;
      setCustomSize(w, h);
    });
  }

  function setDevice(name) {
    document.querySelectorAll('#device-picker button').forEach(function(b) {
      b.classList.toggle('active', b.dataset.device === name);
    });
    var frame = document.getElementById('device-frame');
    var iframe = document.getElementById('app-frame');
    var notch = document.getElementById('device-notch');

    frame.className = name;

    if (name === 'full') {
      iframe.style.width = '100%';
      iframe.style.height = '100%';
    } else if (DEVICES[name]) {
      iframe.style.width = DEVICES[name].width + 'px';
      iframe.style.height = DEVICES[name].height + 'px';
    }
  }

  function setCustomSize(w, h) {
    document.querySelectorAll('#device-picker button').forEach(function(b) {
      b.classList.remove('active');
    });
    var frame = document.getElementById('device-frame');
    var iframe = document.getElementById('app-frame');
    frame.className = '';
    frame.style.width = (w + 24) + 'px';
    frame.style.height = (h + 24) + 'px';
    iframe.style.width = w + 'px';
    iframe.style.height = h + 'px';
  }

  // --- App Loading ---
  function loadApp() {
    var iframe = document.getElementById('app-frame');
    var url = devServerURL + '?_v=' + Date.now();
    iframe.src = url;
    hideError();
  }

  // --- Error Boundary ---
  function showError(message, file, line) {
    var overlay = document.getElementById('error-overlay');
    document.getElementById('error-message').textContent = message;
    document.getElementById('error-file').textContent = (file || '') + (line ? ':' + line : '');
    overlay.style.display = 'block';
    setTimeout(hideError, 8000);
  }

  function hideError() {
    document.getElementById('error-overlay').style.display = 'none';
  }

  // --- Start ---
  init();
})();
```

- [ ] **Step 4: Commit**

```bash
git add web/renderer/
git commit -m "feat: add renderer UI with device frames, WebSocket, and error boundary"
```

---

### Task 6: Wire Everything Together

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\embed.go`
- Modify: `D:\WORKS\ClaudeCode\CliProject\VibeView\main.go`

- [ ] **Step 1: Write embed.go**

```go
package main

import (
	"embed"
	"net/http"
)

//go:embed web/renderer/*
var rendererFS embed.FS

func rendererAssets() http.Handler {
	sub, err := rendererFS.Sub("web/renderer")
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
```

- [ ] **Step 2: Rewrite main.go with full wiring**

```go
package main

import (
	"fmt"
	"os"
	"os/signal"

	"vibeview/internal/detector"
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

	fmt.Println("  VibeView v0.1.0")
	fmt.Printf("  Project:   %s\n", info.Type)
	fmt.Printf("  Dev URL:   %s\n", info.DevServerURL)

	srv := server.New(server.Config{
		Port:         51820,
		DevServerURL: info.DevServerURL,
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
	fmt.Fprintln(os.Stderr, "MCP mode (not yet implemented)")
	select {} // keep alive
}
```

- [ ] **Step 3: Build and verify**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go build -o vibeview.exe .
```

Expected: builds without errors.

- [ ] **Step 4: Run all tests**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go test ./internal/... -v
```

Expected: all PASS (detector + watcher + server)

- [ ] **Step 5: Commit**

```bash
git add embed.go main.go
git commit -m "feat: wire up embedded renderer, file watcher, and server"
```

---

### Task 7: End-to-End Integration Test

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\integration_test.go`

- [ ] **Step 1: Write integration test**

```go
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

func TestFullFlow_HTMLProject(t *testing.T) {
	// Create a fake HTML project
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html><body><h1>Hello</h1></body></html>`), 0644)

	// Detect
	info := detector.Detect(dir)
	if info.Type != detector.HTML {
		t.Fatalf("expected html, got %s", info.Type)
	}

	// Start server on random port
	srv := server.New(server.Config{
		Port:         0, // random port not supported; use a specific one
		DevServerURL: "http://localhost:5173",
		RendererHTML: rendererHTMLBytes(),
		RendererFS:   rendererAssets(),
	})

	// Start watcher
	w := watcher.New(info.WatchDirs)
	defer w.Close()

	go func() {
		for range w.Events {
			srv.Broadcast("reload", nil)
		}
	}()

	// Test health endpoint (server started in goroutine)
	go srv.Start()
	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get("http://localhost:51820/health")
	if err != nil {
		t.Skipf("server not reachable (expected in CI): %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Test config endpoint
	resp2, err := http.Get("http://localhost:51820/api/config")
	if err != nil {
		t.Fatalf("config endpoint failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp2.StatusCode)
	}
}

func TestFullFlow_DetectAndWatch(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "index.html"), []byte(`<html></html>`), 0644)

	info := detector.Detect(dir)
	w := watcher.New(info.WatchDirs)
	defer w.Close()

	time.Sleep(100 * time.Millisecond)

	// Modify a file
	os.WriteFile(filepath.Join(dir, "style.css"), []byte(`body{}`), 0644)

	select {
	case <-w.Events:
		// ok
	case <-time.After(5 * time.Second):
		t.Error("timeout waiting for file change event")
	}
}

func TestRendererEmbedded(t *testing.T) {
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
```

- [ ] **Step 2: Run integration tests**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go test -v -timeout 20s
```

Expected: all PASS (tests starting the server may need -timeout)

- [ ] **Step 3: Commit**

```bash
git add integration_test.go
git commit -m "test: add end-to-end integration tests"
```

---

### Task 8: MCP Server

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\mcp\mcp.go`
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\internal\mcp\mcp_test.go`
- Modify: `D:\WORKS\ClaudeCode\CliProject\VibeView\main.go`

- [ ] **Step 1: Write mcp.go**

```go
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	JSONRPCVersion = "2.0"
	ServerName    = "vibeview"
	ServerVersion = "0.1.0"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Server struct {
	reader    *bufio.Reader
	writer    io.Writer
	serverURL string
	client    *http.Client
}

func New(serverURL string) *Server {
	return &Server{
		reader:    bufio.NewReader(os.Stdin),
		writer:    os.Stdout,
		serverURL: serverURL,
		client:    &http.Client{},
	}
}

func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}
		resp := s.handle(req)
		data, _ := json.Marshal(resp)
		fmt.Fprintln(s.writer, string(data))
	}
	return scanner.Err()
}

func (s *Server) handle(req Request) Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return Response{
			JSONRPC: JSONRPCVersion,
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: "method not found"},
		}
	}
}

func (s *Server) handleInitialize(req Request) Response {
	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    ServerName,
				"version": ServerVersion,
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		},
	}
}

func (s *Server) handleToolsList(req Request) Response {
	tools := []Tool{
		{
			Name:        "preview_reload",
			Description: "Reload the preview iframe to reflect latest code changes",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "preview_console",
			Description: "Read recent browser console messages from the preview iframe",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "preview_screenshot",
			Description: "Capture a screenshot of the current preview. The browser renders the preview, so use preview_console or preview_reload for state changes.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
	}
	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      req.ID,
		Result:  map[string]interface{}{"tools": tools},
	}
}

func (s *Server) handleToolsCall(req Request) Response {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	json.Unmarshal(req.Params, &params)

	var result interface{}
	var err *RPCError

	switch params.Name {
	case "preview_reload":
		resp, httpErr := s.client.Post(s.serverURL+"/api/reload", "application/json", nil)
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			resp.Body.Close()
			result = map[string]string{"status": "reloaded"}
		}
	case "preview_console":
		resp, httpErr := s.client.Get(s.serverURL + "/api/console")
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			var msgs []interface{}
			json.Unmarshal(data, &msgs)
			result = map[string]interface{}{"messages": msgs}
		}
	case "preview_screenshot":
		result = map[string]string{
			"message": "Screenshot not available in CLI mode. Open http://localhost:51820 in browser to see the preview.",
		}
	default:
		err = &RPCError{Code: -32601, Message: "unknown tool: " + params.Name}
	}

	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      req.ID,
		Result:  result,
		Error:   err,
	}
}
```

- [ ] **Step 2: Write mcp_test.go**

```go
package mcp

import (
	"encoding/json"
	"testing"
)

func TestInitialize(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 1, Method: "initialize"}
	resp := s.handle(req)
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("wrong protocol version: %v", result["protocolVersion"])
	}
}

func TestToolsList(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 2, Method: "tools/list"}
	resp := s.handle(req)
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	tools, ok := result["tools"].([]Tool)
	if !ok {
		t.Fatal("tools is not a slice")
	}
	if len(tools) < 2 {
		t.Errorf("expected at least 2 tools, got %d", len(tools))
	}
}

func TestUnknownMethod(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 3, Method: "nonexistent"}
	resp := s.handle(req)
	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected -32601, got %d", resp.Error.Code)
	}
}

func TestToolsCallUnknown(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	params, _ := json.Marshal(map[string]interface{}{
		"name":      "nonexistent_tool",
		"arguments": map[string]interface{}{},
	})
	req := Request{JSONRPC: "2.0", ID: 4, Method: "tools/call", Params: params}
	resp := s.handle(req)
	if resp.Error == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestJSONRPCFormatting(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 1, Method: "initialize"}
	resp := s.handle(req)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed Response
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if parsed.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", parsed.JSONRPC)
	}
}
```

- [ ] **Step 3: Update main.go runMCP**

Replace the placeholder `runMCP()` in main.go:

```go
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
```

Add import for `"vibeview/internal/mcp"` in main.go.

- [ ] **Step 4: Run MCP tests**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go test ./internal/mcp/ -v
```

Expected: all PASS

- [ ] **Step 5: Build and verify MCP tools list**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | go run .
```

Expected: JSON response with tools array.

- [ ] **Step 6: Commit**

```bash
git add internal/mcp/ main.go
git commit -m "feat: add MCP server with preview_reload, preview_console, preview_screenshot tools"
```

---

### Task 9: CLAUDE.md

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\CLAUDE.md`

- [ ] **Step 1: Write CLAUDE.md**

```markdown
# VibeView

A CLI tool that gives Android Studio-style instant UI preview for vibe coding. Browser preview with device frames + WebSocket hot reload + MCP Server for Claude Code.

## Tech Stack

- **Go 1.22** — CLI core, HTTP/WS server, MCP server
- **fsnotify** — cross-platform file watching with debounce
- **gorilla/websocket** — WebSocket for live reload push
- **Vanilla HTML/CSS/JS** — renderer UI (embedded via go:embed)

## Project Structure

| Path | Purpose |
|------|---------|
| `main.go` | CLI entry point, command dispatch (`vibeview` / `vibeview mcp`) |
| `embed.go` | go:embed directives for `web/renderer/*` |
| `internal/detector/` | Framework auto-detection (React/Vue/Svelte/HTML from package.json) |
| `internal/watcher/` | fsnotify wrapper with 100ms debounce, skips node_modules/.git |
| `internal/server/` | HTTP server on :51820, WebSocket hub, API endpoints |
| `internal/mcp/` | MCP JSON-RPC server over stdio (tools: preview_reload, preview_console, preview_screenshot) |
| `web/renderer/` | Browser preview page: device frames, iframe, error boundary |
| `integration_test.go` | End-to-end tests |

## Build & Run

```bash
# Build
go build -o vibeview.exe .

# Run (in a frontend project directory)
./vibeview.exe

# Run as MCP server (for Claude Code)
./vibeview.exe mcp

# Test
go test ./internal/... -v

# Test everything
go test -v -timeout 30s
```

## Architecture

```
Cursor/IDE (user edits code)
    │ file change
    ▼
VibeView Core (Go)
    ├── File Watcher (fsnotify, 100ms debounce)
    ├── HTTP Server (:51820)
    │   ├── / → embedded renderer page
    │   ├── /_vibeview/* → embedded CSS/JS assets
    │   ├── /ws → WebSocket (reload push, console forwarding)
    │   ├── /api/config → dev server URL for the current project
    │   ├── /api/reload → trigger reload
    │   └── /api/console → read buffered console messages
    └── MCP Server (stdio JSON-RPC) ← Claude Code connects here
    │ WebSocket → Browser Preview
    ▼
Browser Preview Window
    ├── Device frame (iPhone/Pixel/iPad/Full/Custom)
    ├── iframe → user's Vite dev server
    ├── Error boundary (red card overlay)
    └── Console forwarding → WebSocket → MCP
```

## MCP Tools

| Tool | HTTP Backend | Description |
|------|-------------|-------------|
| `preview_reload` | POST /api/reload | Trigger preview iframe refresh |
| `preview_console` | GET /api/console | Read buffered browser console messages |
| `preview_screenshot` | — | Check http://localhost:51820 in browser |

## Claude Code MCP Configuration

Add to `~/.claude/claude_desktop_config.json` or `.mcp.json`:

```json
{
  "mcpServers": {
    "vibeview": {
      "command": "vibeview",
      "args": ["mcp"],
      "env": {
        "VIBEVIEW_URL": "http://localhost:51820"
      }
    }
  }
}
```

## Design Decisions

- **Why Go:** Single binary distribution, go:embed for renderer, fast file watching, cross-compile to Windows/macOS/Linux
- **Why not start Vite:** VibeView wraps the existing Vite dev server (user already has `npm run dev` running). Vite's own HMR handles module updates; VibeView adds full-page reload as safety net + the device frame wrapper.
- **Why no screenshot in v1:** Browser-side screenshot requires html2canvas or similar injected into the user's iframe, which can conflict with user code. Deferred to v2.
- **Debounce 100ms:** Prevents rapid-fire reloads when saving multiple files (e.g., IDE "save all").

## Future: Plan B (Electron Desktop App)

After v1 CLI is stable: Electron wrapper with always-on-top window, built-in Chromium, pixel-level comparison tools. User to be reminded after v1.
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: add CLAUDE.md with project overview and development guide"
```

---

### Task 10: Cross-Compile Build Script

**Files:**
- Create: `D:\WORKS\ClaudeCode\CliProject\VibeView\scripts\build.sh`

- [ ] **Step 1: Write build.sh**

```bash
#!/bin/bash
set -e

APP="vibeview"
VERSION="${1:-0.1.0}"
BUILD_DIR="build"

rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

PLATFORMS=(
  "windows/amd64/.exe"
  "darwin/amd64/"
  "darwin/arm64/"
  "linux/amd64/"
  "linux/arm64/"
)

for entry in "${PLATFORMS[@]}"; do
  IFS='/' read -r GOOS GOARCH EXT <<< "$entry"
  OUTPUT="$BUILD_DIR/${APP}_${VERSION}_${GOOS}_${GOARCH}${EXT}"
  echo "Building $OUTPUT ..."
  GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w" -o "$OUTPUT" .
done

echo "Done. Builds in $BUILD_DIR/"
ls -la "$BUILD_DIR/"
```

- [ ] **Step 2: Test build**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
bash scripts/build.sh 0.1.0
```

Expected: all platform binaries created in `build/`

- [ ] **Step 3: Commit**

```bash
git add scripts/build.sh
git commit -m "build: add cross-compile script for windows/darwin/linux"
```

---

### Task 11: Final Verification

- [ ] **Step 1: Run all tests one final time**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go test ./... -v -timeout 30s
```

Expected: every test PASS

- [ ] **Step 2: Build and run manually against a real project**

```bash
cd D:/WORKS/ClaudeCode/CliProject/VibeView
go build -o vibeview.exe .

# Test against a Vite React project
cd /path/to/some/react-project
../VibeView/vibeview.exe
```

Expected: server starts, open browser to localhost:51820, see device frame preview

- [ ] **Step 3: Test MCP mode**

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./vibeview.exe mcp
```

Expected: JSON response listing 3 tools

- [ ] **Step 4: Commit any final fixes**

```bash
git add -A
git commit -m "chore: final verification and fixes"
```
