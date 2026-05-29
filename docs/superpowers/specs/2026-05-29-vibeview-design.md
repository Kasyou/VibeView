# VibeView — Design Spec

## Overview

**VibeView** is a CLI tool that provides Android Studio-style instant UI preview for vibe coding. Write code in your editor, see the UI update instantly in a browser preview window. Doubles as an MCP Server so Claude Code can "see" the rendered output and self-correct UI issues.

**Tagline:** "See what your AI codes."

**Target experience:** Like Android Studio's layout preview — code on one side, live rendering on the other. No manual refresh. No window switching.

---

## Core Architecture

```
┌──────────────────────────────┐
│         Cursor / IDE          │
│  User edits React/Vue/HTML    │
└──────────────┬───────────────┘
               │ file change
               ▼
┌──────────────────────────────┐
│        VibeView Core (Go)     │
│                              │
│  • fsnotify file watcher      │──▶ MCP Server (Claude Code)
│  • WebSocket server           │
│  • Static asset server        │
│  • Framework auto-detection   │
└──────────────┬───────────────┘
               │ WebSocket push + HTTP
               ▼
┌──────────────────────────────┐
│     Browser Preview Window    │
│                              │
│  • Device frame (iPhone etc.) │
│  • iframe sandbox for user app│
│  • Error boundary (no WSOD)   │
│  • HMR (hot module replacement)│
└──────────────────────────────┘
```

### Components

| Component | Technology | Responsibility |
|-----------|-----------|----------------|
| **Core** | Go (fsnotify + gorilla/websocket) | File watching, WebSocket push, MCP protocol, HTTP server |
| **Renderer** | Pure HTML/JS (embedded via go:embed) | Preview window UI: device frame, iframe sandbox, error boundary, console forwarding |
| **MCP Server** | Go stdlib JSON-RPC | Expose preview tools to Claude Code |

### Why Go

- Compiles to single binary (no Node.js dependency)
- `go:embed` bundles the renderer directly into the binary
- fsnotify is reliable on Windows, macOS, Linux
- Cross-compilation trivial: `GOOS=windows GOARCH=amd64 go build`

---

## Features

### Developer Experience (vibeview command)

| Feature | Description |
|---------|-------------|
| One-command start | `vibeview` in any frontend project directory |
| Framework auto-detect | Detects React (Vite), Vue (Vite), Svelte (Vite), plain HTML |
| Instant hot reload | WebSocket push on file change, ~10-20ms latency |
| Device frame preview | iPhone 15 Pro (393×852), Pixel 8 (412×915), iPad (744×1133), custom |
| Error boundary | Code errors show red error card instead of white screen, auto-clear on fix |
| Console bridge | Browser console output forwarded to terminal |
| Window memory | Remembers last window position and size |
| No forced on-top | User controls window behavior, drag to second monitor freely |

### AI Experience (vibeview mcp — MCP Server mode)

| Tool | What the AI can do |
|------|-------------------|
| `preview_screenshot` | Capture current preview page as image |
| `preview_inspect` | Query element styles, position, dimensions |
| `preview_console` | Read browser console errors/warnings |
| `preview_diff` | Compare before/after screenshots for unexpected changes |
| `preview_reload` | Trigger refresh after code changes |

### AI Self-Correction Flow

```
Claude Code generates React component
  → calls preview_screenshot
  → detects button overflow
  → auto-fixes CSS
  → calls preview_screenshot again to verify
  → presents final result to user
```

---

## Supported Frameworks (v1)

All Vite-based projects with zero config:
- HTML / plain static
- React (Vite)
- Vue (Vite)
- Svelte (Vite)

Detection logic: check for `vite.config.{js,ts}`, read `package.json` dependencies, fall back to serving static files.

---

## NON-Goals (v1)

- State management debugging (Redux DevTools territory)
- Network request interception (browser DevTools territory)
- Multi-device sync preview
- Electron desktop app (saved for v2 / Plan B)

---

## Distribution

| Channel | Method |
|---------|--------|
| GitHub Releases | Prebuilt binaries (Windows x64, macOS arm64/x64, Linux x64) |
| Windows | `winget install vibeview` (post-launch) |
| macOS | `brew install vibeview` (post-launch) |
| Go users | `go install github.com/<user>/vibeview@latest` |

---

## Future: Plan B (Electron Desktop App)

After v1 CLI is stable and has traction:
- Electron wrapper with always-on-top window option
- Built-in Chromium for richer inspection tools
- Pixel-level comparison, annotation tools
- User to be reminded when v1 is complete

---

## Project Location

`D:\WORKS\ClaudeCode\CliProject\VibeView`
