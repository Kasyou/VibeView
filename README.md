# VibeView

**"See what your AI codes."**

VibeView gives you Android Studio-style instant UI preview for vibe coding. Write code in your editor, see the UI update instantly in a browser preview window. Doubles as an MCP Server so Claude Code can "see" the rendered output and self-correct UI issues.

<p align="center">
  <img src="screenshots/screenshot-iphone.png" alt="VibeView iPhone 15 Pro preview" width="600">
</p>

## Features

- **One-command start** — `vibeview` in any frontend project directory
- **Framework auto-detect** — React (Vite), Vue (Vite), Svelte (Vite), plain HTML
- **Instant hot reload** — WebSocket push on file change, 100ms debounce
- **Device frames** — iPhone 15 Pro, Pixel 8, iPad, Full Width, or custom size
- **Error boundary** — JS errors show red error card instead of white screen, auto-clear on fix
- **Console bridge** — Browser errors forwarded to terminal in real time
- **MCP Server** — 5 tools for Claude Code to see and interact with the preview

<p align="center">
  <img src="screenshots/screenshot-error.png" alt="Error boundary overlay" width="400">
  <img src="screenshots/screenshot-full.png" alt="Full width preview" width="400">
</p>

## Quick Start

### Install

```bash
# Clone and build
git clone https://github.com/Kasyou/VibeView.git
cd VibeView
go build -o vibeview.exe .

# Or with Go install (requires Go 1.20+)
go install github.com/Kasyou/VibeView@latest
```

### Usage

```bash
# Start preview in any frontend project
cd my-react-app
vibeview

# Open http://localhost:51820 in your browser
```

```bash
# Custom port and directory
vibeview --port 3000 --dir ~/my-project

# Show help
vibeview help
```

### With Claude Code (MCP)

Add to `.mcp.json` in your project:

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

Then Claude Code can:

| Tool | What the AI does |
|------|-----------------|
| `preview_screenshot` | Capture a screenshot of the current preview |
| `preview_inspect` | Query element styles, position, dimensions |
| `preview_console` | Read browser console errors/warnings |
| `preview_diff` | Compare before/after screenshots for changes |
| `preview_reload` | Trigger a preview refresh |

### AI Self-Correction Flow

```
Claude generates React component
  → preview_screenshot
  → detects button overflow
  → auto-fixes CSS
  → preview_screenshot again to verify
  → presents final result to user
```

## Architecture

```
Cursor/IDE (user edits code)
    │ file change
    ▼
VibeView Core (Go binary, ~8MB)
    ├── File Watcher (fsnotify, 100ms debounce)
    ├── HTTP Server (:51820)
    │   ├── /        → embedded preview page
    │   ├── /_app/*  → serves project files (HTML mode)
    │   ├── /ws      → WebSocket (reload, console, screenshot, inspect)
    │   └── /api/*   → REST endpoints
    └── MCP Server (stdio JSON-RPC)
    │ WebSocket → Browser Preview
    ▼
Browser Preview Window
    ├── Device frame (iPhone/Pixel/iPad/Full/Custom)
    ├── iframe → user's project (Vite dev server or local files)
    ├── Error boundary (red card overlay, auto-dismiss)
    └── Console forwarding → WebSocket → terminal
```

## Supported Frameworks

| Framework | Detection | HMR |
|-----------|-----------|-----|
| React (Vite) | `package.json` + `vite.config.*` | Vite built-in |
| Vue (Vite) | `package.json` + `vite.config.*` | Vite built-in |
| Svelte (Vite) | `package.json` + `vite.config.*` | Vite built-in |
| Plain HTML | `index.html` exists | VibeView forced reload |

## Tech Stack

- **Go 1.20+** — Single binary, `go:embed` for renderer
- **fsnotify** — Cross-platform file watching
- **gorilla/websocket** — WebSocket for live reload
- **Vanilla HTML/CSS/JS** — Zero-dependency renderer UI, embedded in binary

## Build from Source

```bash
git clone https://github.com/Kasyou/VibeView.git
cd VibeView
go build -ldflags="-s -w" -o vibeview .

# Cross-compile for all platforms
bash scripts/build.sh 0.1.0
```

## License

MIT
