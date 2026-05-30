---
name: vibeview
description: Use when building UIs, editing frontend code (React/Vue/Svelte/HTML/CSS), verifying visual changes, debugging layout issues, or checking the rendered output of code changes. Provides live preview, element inspection, console error reading, and screenshot comparison.
---

## Overview

You have access to a live browser preview of the user's frontend project. The VibeView preview server runs at `http://localhost:51820` and renders the project inside device frames (iPhone/Pixel/iPad/Full). You can see the preview, inspect elements, read console errors, take screenshots, and compare visual changes — just like Android Studio's layout preview.

## When to Use These Tools

**ALWAYS use VibeView tools when:**
- Writing or editing any HTML, CSS, JSX, TSX, Vue, or Svelte component
- The user asks you to build UI, style something, or fix layout
- Verifying that your code changes produced the expected visual result
- Debugging why something doesn't look right
- The user mentions "preview", "see what it looks like", "check the UI"

**The workflow for every UI change:**
1. Make the code change
2. Run `preview_reload` to refresh the preview
3. Run `preview_screenshot` to see the result
4. If something looks wrong, use `preview_inspect` to query elements
5. If errors exist, use `preview_console` to read them
6. Fix and repeat

## Tools

### preview_screenshot
Capture the current preview as an image. Shows the project inside the device frame with toolbar, status, and any error messages.

Use this BEFORE and AFTER every UI change to verify the result.

### preview_inspect
Query an element's position, size, styles, and text content. Requires a CSS selector.

```
preview_inspect("button")        // first button
preview_inspect(".card")         // element with class 'card'
preview_inspect("#app h1")       // h1 inside #app
```

Returns: `{found, tag, text, rect: {x, y, w, h}, display, color, backgroundColor, fontSize, fontWeight, className, id}`

Use this when you need to check if an element is positioned correctly, has the right size, or to debug overflow/alignment issues.

### preview_console
Read recent browser console errors and warnings. Use after making changes to check for runtime errors.

### preview_diff
Compare the current preview screenshot with the previous one. Returns whether visual changes were detected, with both before/after images.

```
flow: screenshot → make change → screenshot → diff
```

Use this to verify that ONLY the intended elements changed.

### preview_reload
Trigger a forced refresh of the preview iframe. Call after saving code changes.

## Responsible Tool Use

- **Batch your checks**: Make all your code changes, then take one screenshot. Don't screenshot after every single line.
- **Inspect, don't guess**: If a screenshot shows a layout issue, use `preview_inspect` to find the exact element and its dimensions before guessing at a fix.
- **Diff to verify**: After fixing a bug, use `preview_diff` to confirm your fix didn't break anything else.
- **Read errors first**: If something looks wrong, check `preview_console` before inspecting. The error message often tells you exactly what's broken.
- **Timeout handling**: If a tool returns a timeout, the user may not have the preview open in a browser. Ask them to open http://localhost:51820.

## Self-Correction Loop

```
1. Generate/edit UI code
2. preview_reload      → refresh the preview
3. preview_screenshot   → see the result
4. [Looks wrong?]
   a. preview_console  → check for errors
   b. preview_inspect  → query the problem element
   c. Fix the code
   d. preview_reload   → refresh again
   e. preview_screenshot → verify fix
5. preview_diff        → confirm no unintended changes
6. Present result to user
```

## Project Setup

VibeView is already configured in this project's MCP settings. The preview server auto-starts on session launch.

If the preview isn't working:
```
vibeview             # start preview server in project directory
```

Open http://localhost:51820 in a browser to see the live preview.
