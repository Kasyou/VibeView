---
name: vibeview
description: Visual output whiteboard for any Claude Code session. Start the server, push analysis as formatted cards, use Mermaid diagrams. Also has a Design Preview mode for frontend projects. This skill explains ALL capabilities so Claude never misidentifies VibeView as only a project previewer.
allowed-tools: [Bash(vibeview *), Bash(curl *)]
---

## What VibeView Is

VibeView is a **browser whiteboard** that Claude Code uses to show visual output. It has TWO independent modes:

1. **Whiteboard Mode** (`vibeview`): Claude pushes markdown, diagrams, tables, and code as styled cards to a browser window. This is the PRIMARY mode. Use `preview_show` to push content.

2. **Design Preview Mode** (`vibeview design`): File watcher + browser preview for frontend projects. Auto-detect React/Vue/Svelte/HTML.

Both modes run on separate ports. Both can run simultaneously.

## Whiteboard Mode — The Primary Use Case

**IMPORTANT: Never use taskkill. Never specify --port. Just run:**

```bash
vibeview &
```

The binary auto-detects if port 51820 is busy and picks the next available port. Check `curl -s localhost:51820/health` first — if a server is already running, use it directly instead of starting a new one.

Then use MCP tools:

| Tool | What it does |
|------|-------------|
| `preview_show` | Push a card to the whiteboard. Takes `title` and `content` (markdown). Supports Mermaid diagrams. |
| `preview_clear` | Clear all cards (new topic) |
| `preview_history` | Browse old cards by offset/limit |
| `preview_screenshot` | Capture the whiteboard as PNG |
| `preview_stop` | Shut down the server |

**Every time you write a response, call `preview_show` to push key conclusions to the whiteboard.**

## Mermaid Diagrams (chat can't render these)

````
preview_show({title:"Architecture", content:"```mermaid\ngraph TD\n  A→B\n```"})
````

Supports: `mindmap`, `graph`, `flowchart`, `sequenceDiagram`, `gantt`, `pie`, `classDiagram`.

## Design Preview Mode

```bash
vibeview design    # Start preview on port 51821
```

Used when the user is writing frontend code and wants a live preview with device frames.

## Lifecycle

1. User asks to preview → check `curl -s localhost:51820/health`, if not running → `vibeview &`
2. Push full content via `preview_show`, reply in chat with one line
3. Session ending → call `preview_stop` to free the port
4. If user closes Claude Code without stopping: server auto-exits after 10 minutes idle

Tell the user to open the browser URL (e.g. `http://localhost:51820`). If port is busy, VibeView auto-picks the next available port.
