---
name: vibeview
description: Visual output whiteboard for any Claude Code session. Start the server, push analysis as formatted cards, use Mermaid diagrams. Also has a Design Preview mode for frontend projects.
allowed-tools: [Bash(vibeview *), Bash(curl *)]
---

## What VibeView Is

A **browser whiteboard** that Claude pushes visual output to. Two modes:

1. **Whiteboard** (`vibeview`): Push markdown/diagrams as cards via `preview_show`. The primary mode.
2. **Design Preview** (`vibeview design`): Live UI preview with device frames.

## How to Start

**NEVER use taskkill. NEVER specify --port. Just run:**

```bash
vibeview &
```

The binary handles everything: auto-picks an available port (51820, 51821, ...), starts the server, prints the URL. You do NOT need to check ports yourself.

## User Says "Restart"

"Restart" means start a NEW instance on the next port. DO NOT kill the existing one:

```bash
vibeview &
```

The binary finds 51820 busy → auto-uses 51821 → prints the new URL. Multiple instances coexist.

## User Says "Start/Open VibeView"

Same thing. Just:

```bash
vibeview &
```

No port check needed. No taskkill. No `--port`. One command.

## How to Stop

```bash
preview_stop
```

This shuts down ONLY the server on the current port. Other instances are untouched.

## MCP Tools

| Tool | Purpose |
|------|---------|
| `preview_show` | Push markdown card (title + content). Supports Mermaid. |
| `preview_clear` | Clear all cards |
| `preview_stop` | Shut down server |
| `preview_history` | Browse old cards |
| `preview_screenshot` | Capture as PNG |

**Push every response to the whiteboard with `preview_show`. Chat reply is one sentence.**

## Mermaid Diagrams

```js
preview_show({title:"Arch", content:"```mermaid\ngraph TD\n  A→B\n```"})
```

Supports: `mindmap`, `graph`, `flowchart`, `sequenceDiagram`, `gantt`, `pie`.
