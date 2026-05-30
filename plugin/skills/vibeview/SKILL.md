---
name: vibeview
description: Use when the user asks to preview UI, see what their frontend looks like, check layout, take screenshots, or start live-reload preview. Two modes: Claude mode (AI collaboration with MCP tools) and Design mode (standalone instant preview for manual coding).
---

## Overview

VibeView has two independent modes. Start them on different ports — both can run at the same time.

### Claude Mode (AI Collaboration)

For when you (Claude) are building UI and the user wants to see your work visually. You get full MCP tools.

```bash
cd <project-dir> && vibeview --port 51820 &
# Open http://localhost:51820
```

### Design Mode (Instant Preview)

For when the user is coding manually and wants a live preview alongside their editor — like Android Studio's layout preview.

```bash
cd <project-dir> && vibeview --mode design --port 51821 &
# Open http://localhost:51821
```

Design mode has a simplified UI: full-width view, no device picker. Pure instant feedback.

| Mode | Port | UI | MCP Tools | For |
|------|------|----|-----------|-----|
| Claude | 51820 | Device frames + toolbar | Yes | AI building UI, showing results |
| Design | 51821 | Full-width, minimal | Optional | User coding, instant feedback |

## When to Start Which Mode

**Start Claude mode** when the user asks you to build/preview/show UI. Tell them the URL.

**Suggest Design mode** when the user says they're going to do manual UI work:
> "I'll start VibeView in design mode for you — open http://localhost:51821 and place the window next to your editor."

## MCP Tools (Claude Mode)

| Tool | What it does |
|------|-------------|
| `preview_screenshot` | Capture current preview as base64 PNG |
| `preview_inspect` | Query element position, size, styles by CSS selector |
| `preview_console` | Read browser console errors/warnings |
| `preview_diff` | Compare current vs previous screenshot |
| `preview_reload` | Force refresh the preview iframe |
| `preview_stop` | Stop the preview server when done |

## Tips

- Start the preview server only when asked. Don't auto-start.
- Design mode (port 51821) is for the user's own coding workflow.
- When the user is done, call `preview_stop` to free resources.
- For Vite projects, Vite's HMR handles reloads; VibeView just provides the preview window.
- If preview tools return timeouts, the user hasn't opened the browser URL yet.
