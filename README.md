# VibeView

**Visual output for Claude Code.** Claude's analysis appears as styled cards in a browser whiteboard — not buried in chat.

<p align="center">
  <img src="screenshots/whiteboard-demo-v0.3.1.png" alt="Claude Whiteboard" width="720">
</p>

## Two Modes

| Mode | Command | Port | Purpose |
|------|---------|------|---------|
| Claude Whiteboard | `vibeview` | 51820 | AI reasoning → visual cards |
| Design Preview | `vibeview design` | 51821 | Real-time UI preview (device frames) |

### Claude Whiteboard

Claude pushes analysis, decisions, and summaries as styled cards. Chat becomes secondary — the whiteboard is where users look.

<p align="center">
  <img src="screenshots/design-demo-v0.3.1.png" alt="Design Preview" width="480">
</p>

### Design Preview

Android Studio-style instant UI preview. File watcher + WebSocket hot reload. React/Vue/Svelte/HTML auto-detection.

## Quick Start

```bash
go install github.com/Kasyou/VibeView@latest

# Claude whiteboard
vibeview
# → Open http://localhost:51820

# Design preview (project UI)
vibeview design --dir ./my-project
# → Open http://localhost:51821
```

## MCP Tools (9)

| Tool | Description |
|------|-------------|
| `preview_show` | Push markdown card to whiteboard |
| `preview_clear` | Clear all cards |
| `preview_history` | Paginated card history |
| `preview_screenshot` | Capture preview as PNG |
| `preview_inspect` | Query element by CSS selector |
| `preview_console` | Read browser errors |
| `preview_diff` | Compare before/after screenshots |
| `preview_reload` | Force refresh |
| `preview_stop` | Shutdown server |

## Claude Code Plugin

Install `plugin/` to `~/.claude/plugins/cache/local/vibeview/0.1.0/` and enable in settings.json. Claude automatically gets the MCP tools and SKILL.md instruction.

## Build

```bash
git clone https://github.com/Kasyou/VibeView.git
cd VibeView
go build -o vibeview .

# Cross-compile
bash scripts/build.sh
```

## Highlights

- Card annotations: `#seq · timestamp`
- Card queue: survives browser disconnect
- Card limit: 30 DOM / unlimited server history
- Chinese UTF-8 support
- 37 tests

## License

MIT
