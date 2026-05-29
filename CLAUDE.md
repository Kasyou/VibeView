# VibeView

A CLI tool that gives Android Studio-style instant UI preview for vibe coding. Browser preview with device frames + WebSocket hot reload + MCP Server for Claude Code.

## Tech Stack

- **Go 1.20+** — CLI core, HTTP/WS server, MCP server
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
    | file change
    v
VibeView Core (Go)
    +-- File Watcher (fsnotify, 100ms debounce)
    +-- HTTP Server (:51820)
    |   +-- / -> embedded renderer page
    |   +-- /_vibeview/* -> embedded CSS/JS assets
    |   +-- /ws -> WebSocket (reload push, console forwarding)
    |   +-- /api/config -> dev server URL for the current project
    |   +-- /api/reload -> trigger reload
    |   +-- /api/console -> read buffered console messages
    +-- MCP Server (stdio JSON-RPC) <- Claude Code connects here
    | WebSocket -> Browser Preview
    v
Browser Preview Window
    +-- Device frame (iPhone/Pixel/iPad/Full/Custom)
    +-- iframe -> user's Vite dev server
    +-- Error boundary (red card overlay)
    +-- Console forwarding -> WebSocket -> MCP
```

## MCP Tools

| Tool | HTTP Backend | Description |
|------|-------------|-------------|
| `preview_reload` | POST /api/reload | Trigger preview iframe refresh |
| `preview_console` | GET /api/console | Read buffered browser console messages |
| `preview_screenshot` | — | Tells AI to open http://localhost:51820 in browser |

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

- **Why Go:** Single binary distribution, go:embed for renderer, fast file watching, cross-compile
- **Why not start Vite:** VibeView wraps existing Vite dev server. Vite's HMR handles module updates; VibeView adds full-page reload + device frame wrapper.
- **Why no screenshot in v1:** Browser-side screenshot requires html2canvas injected into user's iframe, can conflict with user code. Deferred to v2.
- **Debounce 100ms:** Prevents rapid-fire reloads when saving multiple files.

## Future: Plan B (Electron Desktop App)

After v1 CLI is stable: Electron wrapper with always-on-top window, built-in Chromium, pixel-level comparison tools.
