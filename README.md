# VibeView

[中文](README_CN.md)

**Visual output for Claude Code.** Mermaid diagrams, styled cards, code highlighting. CC is text — VV is structured visualization.

<p align="center">
  <img src="screenshots/vv-capability-demo.png" alt="VibeView capability" width="720">
</p>

## Two Modes

| Mode | Command | Port | Purpose |
|------|---------|------|---------|
| Claude Whiteboard | `vibeview` | 51820 | AI reasoning → diagrams + cards |
| Design Preview | `vibeview design` | 51821 | Code → instant UI preview |

### Design Preview

<p align="center">
  <img src="screenshots/design-demo-v0.3.1.png" alt="Split view: code + live preview" width="720">
</p>

Cursor-dark theme. Device frames. Hot reload. Auto-detect React/Vue/Svelte/HTML.

## Why VV vs CC

| CC (Chat) Can't | VV (Whiteboard) Can |
|-----------------|---------------------|
| Mind maps | `mindmap` — hierarchical visualization |
| Flowcharts | `graph TD` — decision trees, architecture |
| Sequence diagrams | `sequenceDiagram` — API flows |
| Gantt charts | `gantt` — project timelines |
| Code highlighting | GitHub-dark, multi-language |
| Card annotations | #seq · timestamp on every card |
| History search | `preview_history` paginated |

### Smart Browser

VV detects content type. Diagrams → prompts external browser. Text/tables → Cursor built-in browser.

## Quick Start

```bash
go install github.com/Kasyou/VibeView@latest
vibeview            # Whiteboard on :51820
vibeview design     # Preview on :51821
```

## 9 MCP Tools

| Tool | Description |
|------|-------------|
| `preview_show` | Push markdown + diagrams |
| `preview_clear` | Clear whiteboard |
| `preview_history` | Paginated history |
| `preview_screenshot` | Capture PNG |
| `preview_inspect` | CSS selector query |
| `preview_console` | Browser errors |
| `preview_diff` | Before/after compare |
| `preview_reload` | Force refresh |
| `preview_stop` | Shutdown |

## Build

```bash
git clone https://github.com/Kasyou/VibeView.git
cd VibeView && go build -o vibeview .
# Binary: ~12MB (embedded Mermaid.js)
```

37 tests · Go 1.23+ · Windows/macOS/Linux · MIT
