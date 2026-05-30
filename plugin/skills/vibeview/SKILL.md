---
name: vibeview
description: Two independent browser windows — Claude Whiteboard (visualize AI reasoning) and Design Preview (instant UI feedback). Use when the user asks to preview UI, show ideas visually, or see live-reload of their frontend project.
---

## Overview

VibeView has two completely independent modes. Start them on different ports — both can run at the same time.

### Claude Whiteboard (Visual Thinking)

A browser window that renders your markdown as styled cards. Use it to **visualize your reasoning** — show architecture decisions, comparison tables, code snippets, process flows. The user sees your thinking in real time.

```bash
cd <project-dir> && vibeview --port 51820 &
# Open http://localhost:51820
```

**Always-on-top:** run with `--ontop` for PowerShell commands to pin the window.

### Design Preview (Instant UI Feedback)

A browser window with device frames that shows the user's project in real time. File changes trigger instant reload. Pure tool, no MCP clutter.

```bash
cd <project-dir> && vibeview --mode design --port 51821 &
# Open http://localhost:51821
```

## MCP Tools (Claude Whiteboard Mode)

| Tool | Description |
|------|-------------|
| `preview_show` | Push markdown content to the whiteboard — renders as a styled card |
| `preview_clear` | Clear all cards from the whiteboard |
| `preview_screenshot` | Capture whiteboard as base64 PNG |
| `preview_console` | Read browser console errors/warnings |
| `preview_stop` | Stop the server when done |

### Using preview_show

Call `preview_show` to push visual content. Supports full markdown:

```
preview_show({
  title: "Architecture Decision",
  content: "## We should use X\n\n- Reason 1\n- Reason 2\n\n| Option | Cost | Speed |"
})
```

Use it to:
- Visualize architecture decisions
- Show comparison tables
- Display code snippets with explanations
- Present conclusions and next steps
- Draw process flows as bullet lists

## When to Use

**Claude Whiteboard** — when you want to SHOW the user your reasoning, not just tell them
**Design Preview** — when the user is coding and wants instant visual feedback

Start only when asked. Don't auto-start.
