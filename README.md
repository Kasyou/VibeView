# VibeView

[中文](README.zh-CN.md)

**A browser window that shows what Claude Code is building — in real time, with diagrams.**

Claude Code is great at writing code, but its output is trapped in a chat window. VibeView gives Claude a place to **show** you what it's doing: architecture diagrams rendered as SVG, comparison tables with syntax highlighting, structured analysis cards with timestamps. Things that don't fit in a chat bubble.

<p align="center">
  <img src="screenshots/vv-capability-demo.png" alt="VibeView whiteboard with mind map and comparison table" width="720">
</p>

## What It Does

When you work with Claude Code, Claude explains things in text. VibeView takes that text and renders it as **visual output** in a separate browser tab:

- **Mind maps and flowcharts** — Claude writes Mermaid syntax, VibeView renders it as SVG diagrams that a chat window cannot display
- **Structured comparison tables** — side-by-side comparisons with dark theme styling, readable at a glance
- **Syntax-highlighted code blocks** — GitHub-dark theme, multi-language, scrollable
- **Timestamped analysis cards** — each card gets a sequence number and timestamp, forming a browsable history

## Why Not Just Read the Chat?

Chat is linear and ephemeral. Scroll up, lose context. VibeView gives you:

1. **Persistence** — every analysis stays as a card with #ID and timestamp. Use `preview_history` to search through old cards
2. **Diagrams** — Mermaid mind maps and flowcharts that literally cannot render in a chat window
3. **Scanability** — structured tables and headings let you find information at a glance without re-reading paragraphs
4. **Separation of concerns** — code in your editor, preview on the right, Claude's reasoning on the whiteboard. Each has its own space

## Two Modes

### Whiteboard Mode (`vibeview`)

Claude pushes analysis, decisions, and summaries to a browser whiteboard. Port 51820.

```
User: "Should we use Redis or Kafka for our event system?"
Claude: [analyzes requirements, compares options]
         → preview_show with comparison table + architecture diagram
         → Whiteboard shows: formatted table + Mermaid flowchart
```

### Design Preview Mode (`vibeview design`)

Real-time UI preview next to your code editor. File watcher detects changes and reloads instantly. iPhone/Pixel/iPad device frames. Auto-detects React, Vue, Svelte, or plain HTML. Port 51821.

<p align="center">
  <img src="screenshots/design-demo-v0.3.1.png" alt="Code editor and live preview side by side" width="720">
</p>

## Quick Start

```bash
go install github.com/Kasyou/VibeView@latest

# Start whiteboard
vibeview
# Open http://localhost:51820

# Start design preview
vibeview design
# Open http://localhost:51821
```

## MCP Tools for Claude Code

VibeView registers 9 tools that Claude can call during a conversation:

| Tool | What Claude Can Do |
|------|-------------------|
| `preview_show` | Push analysis to the whiteboard as a formatted card |
| `preview_clear` | Clear the whiteboard for a new topic |
| `preview_history` | Search through previous cards by offset and limit |
| `preview_screenshot` | Capture the current whiteboard as a PNG image |
| `preview_inspect` | Query DOM elements by CSS selector |
| `preview_console` | Read browser console errors and warnings |
| `preview_diff` | Compare two screenshots for visual changes |
| `preview_reload` | Force-refresh the preview iframe |
| `preview_stop` | Shut down the server when done |

## Build from Source

```bash
git clone https://github.com/Kasyou/VibeView.git
cd VibeView
go build -o vibeview .

# Binary includes embedded Mermaid.js (~12MB)
# 37 tests passing, Go 1.23+
```

## License

MIT
