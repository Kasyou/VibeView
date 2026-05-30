---
name: vibeview
description: Use when the user asks to preview UI, see what their frontend looks like, check layout, take screenshots of a running project, or start live-reload preview. Provides browser preview with device frames, element inspection, console forwarding, and screenshot comparison.
---

## What This Is

VibeView gives you a live browser preview of the user's frontend project. The user edits code, you (or the file watcher) trigger a reload, and the browser shows the updated UI inside an iPhone/Pixel/iPad device frame.

**The preview server runs on demand — the user asks for it, you start it.**

## How to Start

When the user says "preview this" or "show me what it looks like" or "start vibeview":

```bash
# Start the preview server in the background
cd <project-dir> && vibeview &
```

This starts an HTTP/WebSocket server on port 51820. Then tell the user to open http://localhost:51820 in their browser.

Once the preview server is running, these MCP tools become available:

| Tool | What it does |
|------|-------------|
| `preview_screenshot` | Capture current preview as base64 PNG |
| `preview_inspect` | Query element position, size, styles by CSS selector |
| `preview_console` | Read browser console errors/warnings |
| `preview_diff` | Compare current vs previous screenshot |
| `preview_reload` | Force refresh the preview iframe |

## When to Use

- User asks to "preview the project", "see the UI", "show me what it looks like"
- User wants to check if a layout change worked
- User wants to debug a rendering issue
- User asks "what does this look like?"

## When NOT to Use

- User is doing backend work with no UI changes
- User hasn't asked for a preview
- No frontend project is open (no index.html, no vite project)

## Workflow

```
User: "preview this"
  → Start vibeview in background
  → Tell user to open http://localhost:51820
  → preview_screenshot to show current state

User: "the button looks wrong"
  → preview_inspect("button") to check size/position
  → Fix the CSS
  → preview_reload
  → preview_screenshot to verify
  → preview_diff to confirm only the button changed
```

## Tips

- Start the preview server only when asked. Don't auto-start.
- The preview server watches files and auto-reloads. No need to call `preview_reload` for every change.
- For Vite projects, Vite's own HMR handles reloads; VibeView just provides the preview window.
- If preview tools return timeouts, make sure the user has opened http://localhost:51820 and the status shows "live".
